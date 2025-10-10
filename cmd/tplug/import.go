package main

import (
	"context"
	"fmt"
	"github.com/joeblew999/timelinize-plug/internal/runner"
	"github.com/spf13/cobra"
)

func cmdImport() *cobra.Command {
	var source, from, out, tenant string
	c := &cobra.Command{
		Use:   "import",
		Short: "Run a datasource import inline",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := runner.Import(ctx, source, from, out, tenant); err != nil {
				return fmt.Errorf("import failed: %w", err)
			}
			return nil
		},
	}
	c.Flags().StringVar(&source, "source", "", "datasource name (e.g. google_photos, github)")
	_ = c.MarkFlagRequired("source")
	c.Flags().StringVar(&from, "from", "", "input path or URL")
	_ = c.MarkFlagRequired("from")
	c.Flags().StringVar(&out, "out", "./_timeline", "timeline output directory")
	c.Flags().StringVar(&tenant, "tenant", "local", "tenant id")
	return c
}
