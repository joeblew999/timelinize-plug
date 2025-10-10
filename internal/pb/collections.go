package pb

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// EnsureCollections makes sure core collections exist (tokens, secrets, kv).
func EnsureCollections(app *pocketbase.PocketBase) error {
	da := app.Dao()
	if err := ensureTokens(da); err != nil { return fmt.Errorf("ensure tokens: %w", err) }
	return nil
}

func ensureTokens(da *daos.Dao) error {
	const name = "tokens"
	if da.FindCollectionByNameOrId(name) != nil {
		return nil
	}
	col := &models.Collection{
		Name: name,
		Type: models.CollectionTypeBase,
	}
	col.Schema = schema.NewSchema(
		&schema.SchemaField{Name: "provider", Type: "text", Required: true, Options: &schema.TextOptions{Min: 2, Max: 64}},
		&schema.SchemaField{Name: "payload", Type: "json", Required: true},
	)
	// unique index on provider
	col.Indexes = []string{
		"CREATE UNIQUE INDEX idx_tokens_provider ON tokens (provider);",
	}
	return da.SaveCollection(col)
}
