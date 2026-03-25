package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joeblew999/timelinize-plug/internal/nats"
	"github.com/joeblew999/timelinize-plug/internal/runner"
	"github.com/spf13/cobra"
)

func cmdImport() *cobra.Command {
	var source, from, out, tenant string
	c := &cobra.Command{
		Use:   "import",
		Short: "Run a datasource import inline with real-time progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Default tenant
			if tenant == "" {
				tenant = os.Getenv("TPLUG_TENANT_ID")
				if tenant == "" {
					tenant = "local"
				}
			}

			// Start embedded NATS for progress streaming
			ns, _, err := nats.StartEmbedded(ctx, nats.Options{
				Memory: false,
			})
			if err != nil {
				return fmt.Errorf("start NATS: %w", err)
			}
			defer ns.Shutdown()

			// Wire NATS connection into runner
			runner.SetNATS(ns.ClientConn())

			log.Printf("Starting import: source=%s, input=%s, output=%s, tenant=%s", source, from, out, tenant)
			log.Printf("Watch progress at: http://127.0.0.1:12002/events/import?tenant=%s", tenant)
			log.Printf("Or start server with: tplug serve")

			// Run import with progress tracking
			if err := runner.Import(ctx, source, from, out, tenant); err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			log.Printf("Import completed successfully!")
			return nil
		},
	}
	c.Flags().StringVar(&source, "source", "", "datasource name (e.g. google_photos, github)")
	_ = c.MarkFlagRequired("source")
	c.Flags().StringVar(&from, "from", "", "input path or URL")
	_ = c.MarkFlagRequired("from")
	c.Flags().StringVar(&out, "out", "./_timeline", "timeline output directory")
	c.Flags().StringVar(&tenant, "tenant", "", "tenant id (defaults to TPLUG_TENANT_ID or 'local')")
	return c
}
