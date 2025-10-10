package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// FileTokenStore saves tokens per-provider as raw bytes under ~/.tplug/tokens/
// Will be superseded by PocketBase KV in a follow-up.
type FileTokenStore struct {
	Root string
}

func NewFileTokenStore() *FileTokenStore {
	return &FileTokenStore{Root: filepath.Join(os.Getenv("HOME"), ".tplug", "tokens")}
}

func (s *FileTokenStore) Save(ctx context.Context, provider string, data []byte) error {
	if provider == "" { return errors.New("provider is empty") }
	if err := os.MkdirAll(s.Root, 0o755); err != nil { return err }
	return os.WriteFile(filepath.Join(s.Root, provider+".json"), data, 0o600)
}

func (s *FileTokenStore) Load(ctx context.Context, provider string) ([]byte, error) {
	if provider == "" { return nil, errors.New("provider is empty") }
	return os.ReadFile(filepath.Join(s.Root, provider+".json"))
}
