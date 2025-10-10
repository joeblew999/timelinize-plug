package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// PBTokenStore implements TokenStore using embedded PocketBase.
type PBTokenStore struct {
	App *pocketbase.PocketBase
}

func (s *PBTokenStore) Save(ctx context.Context, provider string, data []byte) error {
	if s.App == nil { return errors.New("PB app is nil") }
	da := s.App.Dao()
	rec, err := upsertToken(da, provider, data)
	if err != nil { return err }
	return da.SaveRecord(rec)
}

func (s *PBTokenStore) Load(ctx context.Context, provider string) ([]byte, error) {
	if s.App == nil { return nil, errors.New("PB app is nil") }
	da := s.App.Dao()
	rec, err := findToken(da, provider)
	if err != nil { return nil, err }
	var raw any
	if err := json.Unmarshal([]byte(rec.GetString("payload")), &raw); err == nil {
		return []byte(rec.GetString("payload")), nil
	}
	return []byte(rec.GetString("payload")), nil
}

// --- helpers ---

func upsertToken(da *daos.Dao, provider string, payload []byte) (*models.Record, error) {
	col := da.FindCollectionByNameOrId("tokens")
	if col == nil { return nil, errors.New("tokens collection missing") }
	rec, _ := da.FindFirstRecordByFilter(col.Id, "provider = {:p}", dbx.Params{"p": provider})
	if rec == nil {
		rec = models.NewRecord(col)
		rec.Set("provider", provider)
	}
	rec.Set("payload", string(payload))
	return rec, nil
}

func findToken(da *daos.Dao, provider string) (*models.Record, error) {
	col := da.FindCollectionByNameOrId("tokens")
	if col == nil { return nil, errors.New("tokens collection missing") }
	return da.FindFirstRecordByFilter(col.Id, "provider = {:p}", dbx.Params{"p": provider})
}
