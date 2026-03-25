package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/timelinize-plug/internal/auth"
	"github.com/joeblew999/timelinize-plug/internal/oauth"
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
			if err != nil {
				return fmt.Errorf("start PB: %w", err)
			}

			cfg, err := oauth.Load(app)
			if err != nil {
				return fmt.Errorf("load oauth config: %w", err)
			}

			if err := applyOAuthEnv(provider, cfg); err != nil {
				return err
			}
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

func applyOAuthEnv(provider string, cfg oauth.Config) error {
	setIfEmpty := func(key, val string) {
		if _, ok := os.LookupEnv(key); ok {
			return
		}
		if val == "" {
			return
		}
		_ = os.Setenv(key, val)
	}

	require := func(key, val, providerName string) error {
		if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
			return nil
		}
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("oauth client %s missing %s; configure via web UI or env", providerName, key)
		}
		_ = os.Setenv(key, val)
		return nil
	}

	switch provider {
	case "google":
		if err := require("GOOGLE_CLIENT_ID", cfg.Google.ClientID, "google"); err != nil {
			return err
		}
		if err := require("GOOGLE_CLIENT_SECRET", cfg.Google.ClientSecret, "google"); err != nil {
			return err
		}
		setIfEmpty("GOOGLE_REDIRECT_URL", cfg.Google.RedirectURL)
	case "github":
		if err := require("GITHUB_CLIENT_ID", cfg.GitHub.ClientID, "github"); err != nil {
			return err
		}
		if err := require("GITHUB_CLIENT_SECRET", cfg.GitHub.ClientSecret, "github"); err != nil {
			return err
		}
		setIfEmpty("GITHUB_REDIRECT_URL", cfg.GitHub.RedirectURL)
	default:
		// nothing for other providers yet
	}
	return nil
}
