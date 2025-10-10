package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/pocketbase/pocketbase"
)

// PBTokenStore implements TokenStore using embedded PocketBase,
// but falls back to FileTokenStore until collection is created.
type PBTokenStore struct {
	App  *pocketbase.PocketBase
	File *FileTokenStore
}

func (s *PBTokenStore) Save(ctx context.Context, provider string, data []byte) error {
	if s.App == nil {
		if s.File != nil { return s.File.Save(ctx, provider, data) }
		return errors.New("PB app is nil and no file fallback")
	}
	// TODO: upsert into "tokens" collection
	return fmt.Errorf("PB save not implemented; provider=%s", provider)
}

func (s *PBTokenStore) Load(ctx context.Context, provider string) ([]byte, error) {
	if s.App == nil {
		if s.File != nil { return s.File.Load(ctx, provider) }
		return nil, errors.New("PB app is nil and no file fallback")
	}
	// TODO: lookup from "tokens" collection
	return nil, fmt.Errorf("PB load not implemented; provider=%s", provider)
}
