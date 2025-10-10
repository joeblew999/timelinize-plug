package config

import (
	"errors"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// KV uses a PB collection named "kv" with fields: key(text, unique), value(json).
func EnsureKV(app *pocketbase.PocketBase) error {
	da := app.Dao()
	if da.FindCollectionByNameOrId("kv") != nil { return nil }
	col := &models.Collection{Name: "kv", Type: models.CollectionTypeBase}
	col.Schema.AddField(&models.SchemaField{Name: "key", Type: "text", Required: true})
	col.Schema.AddField(&models.SchemaField{Name: "value", Type: "json", Required: true})
	col.Indexes = []string{"CREATE UNIQUE INDEX idx_kv_key ON kv (key);"}
	return da.SaveCollection(col)
}

func kvSet(da *daos.Dao, key string, raw []byte) error {
	col := da.FindCollectionByNameOrId("kv")
	if col == nil { return errors.New("kv collection missing") }
	rec, _ := da.FindFirstRecordByFilter(col.Id, "key = {:k}", map[string]any{"k": key})
	if rec == nil { rec = models.NewRecord(col); rec.Set("key", key) }
	rec.Set("value", string(raw))
	return da.SaveRecord(rec)
}

func kvGet(da *daos.Dao, key string) ([]byte, error) {
	col := da.FindCollectionByNameOrId("kv")
	if col == nil { return nil, errors.New("kv collection missing") }
	rec, err := da.FindFirstRecordByFilter(col.Id, "key = {:k}", map[string]any{"k": key})
	if err != nil { return nil, err }
	return []byte(rec.GetString("value")), nil
}

// Public helpers
func Set(app *pocketbase.PocketBase, key string, raw []byte) error { return kvSet(app.Dao(), key, raw) }
func Get(app *pocketbase.PocketBase, key string) ([]byte, error) { return kvGet(app.Dao(), key) }
