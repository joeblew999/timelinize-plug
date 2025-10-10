package storage

import (
	"context"
)

// TokenStore abstracts persistence for OAuth tokens (PB KV or local file).
type TokenStore interface {
	Save(ctx context.Context, provider string, data []byte) error
	Load(ctx context.Context, provider string) ([]byte, error)
}
