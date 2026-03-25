package pb

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/migrations/logs"
	"github.com/pocketbase/pocketbase/tools/migrate"

	"github.com/joeblew999/timelinize-plug/internal/config"
	"github.com/joeblew999/timelinize-plug/internal/storage"
)

// Options configures the embedded PocketBase instance.
type Options struct {
	DataDir string // default: $HOME/.tplug/pb
}

// StartEmbedded boots an embedded PocketBase and ensures required collections.
func StartEmbedded(opt Options) (*pocketbase.PocketBase, error) {
	if opt.DataDir == "" {
		opt.DataDir = filepath.Join(".", ".data", "pb")
	}
	if err := os.MkdirAll(opt.DataDir, 0o755); err != nil {
		return nil, err
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: opt.DataDir})

	if err := app.Bootstrap(); err != nil {
		return nil, err
	}

	if err := runMigrations(app); err != nil {
		return nil, err
	}

	if err := EnsureCollections(app); err != nil {
		return nil, err
	}
	if err := config.EnsureKV(app); err != nil {
		return nil, err
	}
    if err := storage.EnsureSecretsCollection(app); err != nil {
        return nil, err
    }

	go func() {
		if err := app.Start(); err != nil {
			log.Printf("pocketbase start: %v", err)
		}
	}()

	return app, nil
}

func runMigrations(app *pocketbase.PocketBase) error {
	type conn struct {
		db   *dbx.DB
		list migrate.MigrationsList
	}

	connections := []conn{
		{db: app.DB(), list: migrations.AppMigrations},
		{db: app.LogsDB(), list: logs.LogsMigrations},
	}

	for _, c := range connections {
		runner, err := migrate.NewRunner(c.db, c.list)
		if err != nil {
			return err
		}
		if _, err := runner.Up(); err != nil {
			return err
		}
	}

	return nil
}
