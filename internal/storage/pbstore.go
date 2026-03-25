package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// ErrTokenNotFound is returned when no token exists for the requested provider.
var ErrTokenNotFound = errors.New("token not found")

// PBTokenStore persists OAuth tokens inside PocketBase.
// A fallback TokenStore can be provided for legacy migrations.
type PBTokenStore struct {
	App      *pocketbase.PocketBase
	Fallback TokenStore
}

func (s *PBTokenStore) Save(ctx context.Context, provider string, data []byte) error {
	if provider == "" {
		return errors.New("provider is empty")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.App == nil {
		if s.Fallback != nil {
			return s.Fallback.Save(ctx, provider, data)
		}
		return errors.New("pocketbase app is nil")
	}

	payload := base64.StdEncoding.EncodeToString(data)
	da := s.App.Dao()
	rec, err := upsertToken(da, provider, payload)
	if err != nil {
		if s.Fallback != nil {
			return s.Fallback.Save(ctx, provider, data)
		}
		return err
	}
	return da.SaveRecord(rec)
}

func (s *PBTokenStore) Load(ctx context.Context, provider string) ([]byte, error) {
	if provider == "" {
		return nil, errors.New("provider is empty")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.App == nil {
		if s.Fallback != nil {
			return s.Fallback.Load(ctx, provider)
		}
		return nil, errors.New("pocketbase app is nil")
	}

	rec, err := findToken(s.App.Dao(), provider)
	if err != nil {
		if s.Fallback != nil && errors.Is(err, ErrTokenNotFound) {
			return s.Fallback.Load(ctx, provider)
		}
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(rec.GetString("payload"))
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	return data, nil
}

func upsertToken(da *daos.Dao, provider string, payload string) (*models.Record, error) {
	col, err := da.FindCollectionByNameOrId("tokens")
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, errors.New("tokens collection missing")
	}
	rec, err := da.FindFirstRecordByFilter(col.Id, "provider = {:p}", dbx.Params{"p": provider})
	if err != nil && rec == nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if rec == nil {
		rec = models.NewRecord(col)
	}
	rec.Set("provider", provider)
	rec.Set("payload", payload)
	return rec, nil
}

func findToken(da *daos.Dao, provider string) (*models.Record, error) {
	col, err := da.FindCollectionByNameOrId("tokens")
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, errors.New("tokens collection missing")
	}
	rec, err := da.FindFirstRecordByFilter(col.Id, "provider = {:p}", dbx.Params{"p": provider})
	if err != nil && rec == nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	if rec == nil {
		return nil, ErrTokenNotFound
	}
	return rec, nil
}
