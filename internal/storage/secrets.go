package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
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
	if existing, err := da.FindCollectionByNameOrId(secretsCollection); err == nil {
		if existing != nil {
			return nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	col := &models.Collection{
		Name: secretsCollection,
		Type: models.CollectionTypeBase,
		Schema: schema.NewSchema(
			&schema.SchemaField{Name: keyField, Type: "text", Required: true},
			&schema.SchemaField{Name: valueField, Type: "text", Required: true},
		),
		Indexes: []string{"CREATE UNIQUE INDEX idx_secrets_key ON secrets (key);"},
	}
	return da.SaveCollection(col)
}

// GetOrCreateAESKey returns a 32-byte key from PB secrets, creating it if missing.
func GetOrCreateAESKey(app *pocketbase.PocketBase) ([]byte, error) {
	da := app.Dao()
	if err := EnsureSecretsCollection(app); err != nil {
		return nil, err
	}
	col, err := da.FindCollectionByNameOrId(secretsCollection)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, errors.New("secrets collection missing")
	}
	rec, err := da.FindFirstRecordByFilter(col.Id, "key = {:k}", dbx.Params{"k": aesKeyName})
	if err != nil && rec == nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if rec == nil {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		rec = models.NewRecord(col)
		rec.Set(keyField, aesKeyName)
		rec.Set(valueField, base64.StdEncoding.EncodeToString(key))
		if err := da.SaveRecord(rec); err != nil {
			return nil, err
		}
		return key, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(rec.GetString(valueField))
	if err != nil {
		return nil, err
	}
	if len(decoded) != 32 {
		return nil, errors.New("invalid aes key length")
	}
	return decoded, nil
}
