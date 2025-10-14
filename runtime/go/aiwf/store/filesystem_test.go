package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFSStorePutGet(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFSStore(Options{Root: dir})
	if err != nil {
		t.Fatalf("NewFSStore: %v", err)
	}

	key := store.Key("workflow", "step", "item", "")
	if err := store.Put(context.Background(), key, []byte("payload")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	data, ok, err := store.Get(context.Background(), key)
	if err != nil || !ok {
		t.Fatalf("Get: %v ok=%v", err, ok)
	}

	if string(data) != "payload" {
		t.Fatalf("unexpected payload: %s", data)
	}
}

func TestFSStoreTTL(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFSStore(Options{Root: dir, TTL: time.Second})
	if err != nil {
		t.Fatalf("NewFSStore: %v", err)
	}

	store.clock = func() time.Time { return time.Unix(100, 0) }

	key := "wf/step/item.json"
	path := filepath.Join(dir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// установить время модификации, чтобы сработал TTL
	if err := os.Chtimes(path, time.Unix(10, 0), time.Unix(10, 0)); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	store.clock = func() time.Time { return time.Unix(200, 0) }

	if _, ok, err := store.Get(context.Background(), key); err != nil || ok {
		t.Fatalf("expected miss due to TTL, err=%v ok=%v", err, ok)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}
}

func TestFSStoreSweep(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFSStore(Options{Root: dir, TTL: time.Second})
	if err != nil {
		t.Fatalf("NewFSStore: %v", err)
	}

	store.clock = func() time.Time { return time.Unix(100, 0) }

	staged := filepath.Join(dir, "wf/step/item.json")
	if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(staged, []byte("data"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chtimes(staged, time.Unix(10, 0), time.Unix(10, 0)); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	store.clock = func() time.Time { return time.Unix(200, 0) }
	if err := store.Sweep(); err != nil {
		t.Fatalf("Sweep: %v", err)
	}

	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}
}

func TestNewFSStoreRequiresRoot(t *testing.T) {
	if _, err := NewFSStore(Options{}); err == nil {
		t.Fatal("expected error without root")
	}
}
