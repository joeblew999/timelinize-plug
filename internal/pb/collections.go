package pb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// EnsureCollections makes sure core collections exist (tokens, secrets, kv).
func EnsureCollections(app *pocketbase.PocketBase) error {
	da := app.Dao()
	if err := ensureTokens(da); err != nil {
		return fmt.Errorf("ensure tokens: %w", err)
	}
	return nil
}

func ensureTokens(da *daos.Dao) error {
	const name = "tokens"
	existing, err := da.FindCollectionByNameOrId(name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if existing != nil {
		return nil
	}
	col := &models.Collection{
		Name: name,
		Type: models.CollectionTypeBase,
	}
	col.Schema = schema.NewSchema(
		&schema.SchemaField{Name: "provider", Type: "text", Required: true},
		&schema.SchemaField{Name: "payload", Type: "text", Required: true},
	)
	// unique index on provider
	col.Indexes = []string{
		"CREATE UNIQUE INDEX idx_tokens_provider ON tokens (provider);",
	}
	return da.SaveCollection(col)
}
