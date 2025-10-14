package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client описывает минимальный набор методов S3, необходимых сторанию.
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// S3Store сохраняет артефакты в S3-совместимом бакете.
type S3Store struct {
	client S3Client
	bucket string
	prefix string
}

// S3Options задаёт конфигурацию клиента S3.
type S3Options struct {
	Client S3Client
	Bucket string
	Prefix string
}

// NewS3Store создаёт S3 store.
func NewS3Store(opts S3Options) (*S3Store, error) {
	if opts.Client == nil {
		return nil, errors.New("store: s3 client is required")
	}
	if opts.Bucket == "" {
		return nil, errors.New("store: s3 bucket is required")
	}

	prefix := opts.Prefix
	if prefix != "" {
		prefix = trimSlashes(prefix)
	}

	return &S3Store{
		client: opts.Client,
		bucket: opts.Bucket,
		prefix: prefix,
	}, nil
}

// Put сохраняет объект в бакете.
func (s *S3Store) Put(ctx context.Context, key string, data []byte) error {
	resolvedKey := s.resolveKey(key)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucket,
		Key:           &resolvedKey,
		Body:          bytesReader(data),
		ContentType:   aws.String("application/json"),
		ContentLength: aws.Int64(int64(len(data))),
		Metadata: map[string]string{
			"aiwf-hash": hashBytes(data),
		},
	})
	var apiErr *types.NoSuchBucket
	if errors.As(err, &apiErr) {
		return fmt.Errorf("store: put object: %w", err)
	}
	return err
}

// Get получает объект из бакета.
func (s *S3Store) Get(ctx context.Context, key string) ([]byte, bool, error) {
	resolvedKey := s.resolveKey(key)
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &resolvedKey,
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// Key генерирует ключ относительного пути.
func (s *S3Store) Key(workflow, step, item, inputHash string) string {
	if inputHash == "" {
		inputHash = hashString(item)
	}
	key := path.Join(workflow, step, fmt.Sprintf("%s.json", inputHash))
	return key
}

func (s *S3Store) resolveKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return path.Join(s.prefix, key)
}

func bytesReader(data []byte) io.ReadSeeker {
	return bytes.NewReader(data)
}

func trimSlashes(p string) string {
	return strings.Trim(p, "/")
}
