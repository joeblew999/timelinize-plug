package storage

import (
	"context"
	"errors"
	"github.com/pocketbase/pocketbase"
)

// PBTokenStore implements TokenStore using an embedded PocketBase.
type PBTokenStore struct {
	App *pocketbase.PocketBase
}

func (s *PBTokenStore) Save(ctx context.Context, provider string, data []byte) error {
	if s.App == nil { return errors.New("PB app is nil") }
	// TODO: create collection "tokens" if missing and upsert by provider
	return errors.New("PBTokenStore.Save not implemented yet")
}

func (s *PBTokenStore) Load(ctx context.Context, provider string) ([]byte, error) {
	if s.App == nil { return nil, errors.New("PB app is nil") }
	// TODO: read token by provider from "tokens" collection
	return nil, errors.New("PBTokenStore.Load not implemented yet")
}
