package storage

import (
	"crypto/rand"
	"errors"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// Secrets collection name and key id
const (
	secretsCollection = "secrets"
	keyField          = "key"
	valueField        = "value"
	aesKeyName        = "aes_gcm_32"
)

// EnsureSecretsCollection ensures a simple key/value secrets collection exists.
func EnsureSecretsCollection(app *pocketbase.PocketBase) error {
	da := app.Dao()
	if da.FindCollectionByNameOrId(secretsCollection) != nil { return nil }
	col := &models.Collection{Name: secretsCollection, Type: models.CollectionTypeBase}
	col.Schema.AddField(&models.SchemaField{Name: keyField, Type: "text", Required: true})
	col.Schema.AddField(&models.SchemaField{Name: valueField, Type: "text", Required: true})
	col.Indexes = []string{"CREATE UNIQUE INDEX idx_secrets_key ON secrets (key);"}
	return da.SaveCollection(col)
}

// GetOrCreateAESKey returns a 32-byte key from PB secrets, creating it if missing.
func GetOrCreateAESKey(app *pocketbase.PocketBase) ([]byte, error) {
	da := app.Dao()
	if err := EnsureSecretsCollection(app); err != nil { return nil, err }
	col := da.FindCollectionByNameOrId(secretsCollection)
	rec, _ := da.FindFirstRecordByFilter(col.Id, "key = {:k}", map[string]any{"k": aesKeyName})
	if rec == nil {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil { return nil, err }
		rec = models.NewRecord(col)
		rec.Set(keyField, aesKeyName)
		rec.Set(valueField, string(key))
		if err := da.SaveRecord(rec); err != nil { return nil, err }
		return key, nil
	}
	val := []byte(rec.GetString(valueField))
	if len(val) != 32 { return nil, errors.New("invalid aes key length") }
	return val, nil
}
