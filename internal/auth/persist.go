package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joeblew999/timelinize-plug/internal/storage"
)

// PersistRawToken is a helper used by provider-specific auth flows.
func PersistRawToken(ctx context.Context, store storage.TokenStore, provider string, raw any) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	return store.Save(ctx, provider, b)
}
