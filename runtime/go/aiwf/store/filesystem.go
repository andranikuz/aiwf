package store

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FSStore сохраняет артефакты на локальной файловой системе.
type FSStore struct {
	root  string
	ttl   time.Duration
	clock func() time.Time
	mu    sync.Mutex
}

// Options задаёт параметры файлового хранилища.
type Options struct {
	Root string
	TTL  time.Duration
}

// NewFSStore создаёт файловый store.
func NewFSStore(opts Options) (*FSStore, error) {
	if opts.Root == "" {
		return nil, errors.New("store: root path is required")
	}

	if err := os.MkdirAll(opts.Root, 0o755); err != nil {
		return nil, fmt.Errorf("store: create root: %w", err)
	}

	ttl := opts.TTL
	if ttl < 0 {
		ttl = 0
	}

	return &FSStore{
		root:  opts.Root,
		ttl:   ttl,
		clock: time.Now,
	}, nil
}

// Put сохраняет бинарный артефакт.
func (s *FSStore) Put(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path := filepath.Join(s.root, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}

// Get возвращает сохранённый артефакт.
func (s *FSStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	path := filepath.Join(s.root, key)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if s.ttl > 0 {
		info, err := os.Stat(path)
		if err == nil {
			if s.clock().Sub(info.ModTime()) > s.ttl {
				_ = os.Remove(path)
				return nil, false, nil
			}
		}
	}

	return data, true, nil
}

// Key генерирует ключ вида workflow/step/item/hash.
func (s *FSStore) Key(workflow, step, item, inputHash string) string {
	if inputHash == "" {
		inputHash = hashString(item)
	}
	return filepath.Join(workflow, step, fmt.Sprintf("%s.json", inputHash))
}

// Sweep удаляет устаревшие файлы по TTL.
func (s *FSStore) Sweep() error {
	if s.ttl == 0 {
		return nil
	}

	cutoff := s.clock().Add(-s.ttl)

	s.mu.Lock()
	defer s.mu.Unlock()

	return filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
				return removeErr
			}
		}
		return nil
	})
}
