package pb

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Options configures the embedded PocketBase instance.
type Options struct {
	DataDir string // default: $HOME/.tplug/pb
}

// StartEmbedded boots an embedded PocketBase (NO CGO via modernc SQLite).
// It runs sync (does not start HTTP). You can register hooks here later.
func StartEmbedded(opt Options) (*pocketbase.PocketBase, error) {
	if opt.DataDir == "" {
		opt.DataDir = filepath.Join(os.Getenv("HOME"), ".tplug", "pb")
	}
	if err := os.MkdirAll(opt.DataDir, 0o755); err != nil {
		return nil, err
	}
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: opt.DataDir,
	})

	// Hook point for migrations/collections (tokens, secrets).
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		log.Printf("PocketBase data dir: %s", opt.DataDir)
		return nil
	})

	// Launch the app without HTTP server (headless).
	go func() {
		if err := app.Start(); err != nil {
			log.Printf("pocketbase start: %v", err)
		}
	}()
	return app, nil
}
