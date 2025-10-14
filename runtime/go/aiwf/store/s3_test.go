package store

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type stubS3 struct {
	put func(context.Context, *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	get func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

func (s *stubS3) PutObject(ctx context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if s.put != nil {
		return s.put(ctx, input)
	}
	return &s3.PutObjectOutput{}, nil
}

func (s *stubS3) GetObject(ctx context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if s.get != nil {
		return s.get(ctx, input)
	}
	return &s3.GetObjectOutput{}, nil
}

func TestS3StorePutGet(t *testing.T) {
	var putCalled bool
	payload := []byte(`{"value":"ok"}`)
	expectedKey := "wf/step/" + hashString("hash") + ".json"

	client := &stubS3{
		put: func(ctx context.Context, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			putCalled = true
			if *input.Key != expectedKey {
				t.Fatalf("unexpected key %s", *input.Key)
			}
			buf, err := io.ReadAll(input.Body)
			if err != nil {
				return nil, err
			}
			if string(buf) != string(payload) {
				t.Fatalf("unexpected payload %s", buf)
			}
			return &s3.PutObjectOutput{}, nil
		},
		get: func(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			if *input.Key != expectedKey {
				t.Fatalf("unexpected get key %s", *input.Key)
			}
			return &s3.GetObjectOutput{Body: io.NopCloser(bytesReader(payload))}, nil
		},
	}

	store, err := NewS3Store(S3Options{Client: client, Bucket: "bucket"})
	if err != nil {
		t.Fatalf("NewS3Store: %v", err)
	}

	key := store.Key("wf", "step", "hash", "")
	if err := store.Put(context.Background(), key, payload); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if !putCalled {
		t.Fatalf("expected put to be called")
	}

	out, ok, err := store.Get(context.Background(), key)
	if err != nil || !ok {
		t.Fatalf("Get: %v ok=%v", err, ok)
	}
	if string(out) != string(payload) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestS3StoreNotFound(t *testing.T) {
	client := &stubS3{
		get: func(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			return nil, &types.NoSuchKey{}
		},
	}

	store, err := NewS3Store(S3Options{Client: client, Bucket: "bucket"})
	if err != nil {
		t.Fatalf("NewS3Store: %v", err)
	}

	if _, ok, err := store.Get(context.Background(), "foo"); err != nil || ok {
		t.Fatalf("expected miss: err=%v ok=%v", err, ok)
	}
}

func TestS3StoreRequiresConfig(t *testing.T) {
	if _, err := NewS3Store(S3Options{}); err == nil {
		t.Fatal("expected error without client and bucket")
	}
	if _, err := NewS3Store(S3Options{Client: &stubS3{}}); err == nil {
		t.Fatal("expected error without bucket")
	}
}

func TestS3StorePropagatesErrors(t *testing.T) {
	expected := errors.New("boom")
	client := &stubS3{
		put: func(ctx context.Context, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			return nil, expected
		},
	}
	store, err := NewS3Store(S3Options{Client: client, Bucket: "bucket"})
	if err != nil {
		t.Fatalf("NewS3Store: %v", err)
	}
	if err := store.Put(context.Background(), "key", []byte("data")); !errors.Is(err, expected) {
		t.Fatalf("expected error propagation, got %v", err)
	}
}
