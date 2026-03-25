package config

import (
	"database/sql"
	"errors"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// EnsureKV guarantees the PocketBase collection used for configuration exists.
func EnsureKV(app *pocketbase.PocketBase) error {
	da := app.Dao()
	if existing, err := da.FindCollectionByNameOrId("kv"); err == nil {
		if existing != nil {
			return nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	col := &models.Collection{
		Name: "kv",
		Type: models.CollectionTypeBase,
		Schema: schema.NewSchema(
			&schema.SchemaField{Name: "key", Type: "text", Required: true},
			&schema.SchemaField{Name: "value", Type: "json", Required: true},
		),
		Indexes: []string{"CREATE UNIQUE INDEX idx_kv_key ON kv (key);"},
	}
	return da.SaveCollection(col)
}

func kvSet(da *daos.Dao, key string, raw []byte) error {
	col, err := da.FindCollectionByNameOrId("kv")
	if err != nil {
		return err
	}
	if col == nil {
		return errors.New("kv collection missing")
	}
	rec, err := da.FindFirstRecordByFilter(col.Id, "key = {:k}", dbx.Params{"k": key})
	if err != nil && rec == nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if rec == nil {
		rec = models.NewRecord(col)
	}
	rec.Set("key", key)
	rec.Set("value", string(raw))
	return da.SaveRecord(rec)
}

func kvGet(da *daos.Dao, key string) ([]byte, error) {
	col, err := da.FindCollectionByNameOrId("kv")
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, errors.New("kv collection missing")
	}
	rec, err := da.FindFirstRecordByFilter(col.Id, "key = {:k}", dbx.Params{"k": key})
	if err != nil && rec == nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("kv entry not found")
		}
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("kv entry not found")
	}
	return []byte(rec.GetString("value")), nil
}

// Set stores a JSON blob under the supplied key.
func Set(app *pocketbase.PocketBase, key string, raw []byte) error { return kvSet(app.Dao(), key, raw) }

// Get retrieves the JSON blob associated with key.
func Get(app *pocketbase.PocketBase, key string) ([]byte, error) { return kvGet(app.Dao(), key) }
