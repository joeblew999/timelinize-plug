package main

import (
	"context"
	"fmt"
	"github.com/joeblew999/timelinize-plug/internal/auth"
	"github.com/joeblew999/timelinize-plug/internal/storage"
	"github.com/spf13/cobra"
)

func cmdAuth() *cobra.Command {
	var provider string
	c := &cobra.Command{
		Use:   "auth",
		Short: "Run OAuth flow and persist token",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store := &storage.PBTokenStore{} // will bind PB instance in serve mode; file fallback not needed here
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
	return c
}
