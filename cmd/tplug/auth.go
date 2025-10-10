package main

import (
	"context"
	"fmt"
	"github.com/joeblew999/timelinize-plug/internal/auth"
	"github.com/spf13/cobra"
)

func cmdAuth() *cobra.Command {
	var provider string
	c := &cobra.Command{
		Use:   "auth",
		Short: "Run OAuth flow and persist token",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := auth.Authenticate(ctx, provider); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
			return nil
		},
	}
	c.Flags().StringVar(&provider, "provider", "google", "google|github|apple")
	return c
}
