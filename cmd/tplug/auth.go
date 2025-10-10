package main

import (
	"context"
	"fmt"

	"github.com/joeblew999/timelinize-plug/internal/auth"
	"github.com/joeblew999/timelinize-plug/internal/pb"
	"github.com/joeblew999/timelinize-plug/internal/storage"
	"github.com/spf13/cobra"
)

func cmdAuth() *cobra.Command {
	var provider, tenant string
	c := &cobra.Command{
		Use:   "auth",
		Short: "Run OAuth flow and persist token (PB encrypted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			app, err := pb.StartEmbedded(pb.Options{})
			if err != nil { return fmt.Errorf("start PB: %w", err) }
			// encrypted PB token store
			store := &storage.PBTokenStoreEnc{App: app}
			switch provider {
			case "google":
				return auth.AuthGoogle(ctx, store)
			case "github":
				return auth.AuthGitHub(ctx, store)
			default:
				return fmt.Errorf("unsupported provider: %s", provider)
			}
		},
	}
	c.Flags().StringVar(&provider, "provider", "google", "google|github")
	c.Flags().StringVar(&tenant, "tenant", "local", "tenant id")
	return c
}
