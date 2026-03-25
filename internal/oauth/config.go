package oauth

import (
	"encoding/json"
	"strings"

	"github.com/joeblew999/timelinize-plug/internal/config"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models/settings"
)

const kvKey = "oauth2/clients"

// Client holds OAuth client metadata.
type Client struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// Config aggregates all supported OAuth clients.
type Config struct {
	Google Client `json:"google"`
	GitHub Client `json:"github"`
}

// Default returns a config with sensible defaults (localhost redirect).
func Default() Config {
	redirect := "http://127.0.0.1:8008/oauth2-redirect"
	return Config{
		Google: Client{RedirectURL: redirect},
		GitHub: Client{RedirectURL: redirect},
	}
}

// Sanitize trims fields and applies defaults.
func Sanitize(c *Client) {
	if c == nil {
		return
	}
	c.ClientID = strings.TrimSpace(c.ClientID)
	c.ClientSecret = strings.TrimSpace(c.ClientSecret)
	if strings.TrimSpace(c.RedirectURL) == "" {
		c.RedirectURL = "http://127.0.0.1:8008/oauth2-redirect"
	}
	c.Enabled = c.ClientID != "" && c.ClientSecret != ""
}

// Load pulls the OAuth configuration from PocketBase KV storage.
func Load(app *pocketbase.PocketBase) (Config, error) {
	cfg := Default()
	if app == nil {
		Sanitize(&cfg.Google)
		Sanitize(&cfg.GitHub)
		return cfg, nil
	}
	if data, err := config.Get(app, kvKey); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			cfg = Default()
		}
	}
	Sanitize(&cfg.Google)
	Sanitize(&cfg.GitHub)
	applyProvidersFromPocketBase(app, &cfg)
	Sanitize(&cfg.Google)
	Sanitize(&cfg.GitHub)
	return cfg, nil
}

// Save persists the OAuth configuration to PocketBase KV storage.
func Save(app *pocketbase.PocketBase, cfg Config) error {
	if app == nil {
		return nil
	}
	Sanitize(&cfg.Google)
	Sanitize(&cfg.GitHub)
	if err := persistPocketBaseProviders(app, cfg); err != nil {
		return err
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return config.Set(app, kvKey, raw)
}

func applyProvidersFromPocketBase(app *pocketbase.PocketBase, cfg *Config) {
	if app == nil || cfg == nil {
		return
	}
	settingsClone, err := app.Settings().Clone()
	if err != nil || settingsClone == nil {
		return
	}
	mergeClientFromProvider(&cfg.Google, settingsClone.GoogleAuth)
	mergeClientFromProvider(&cfg.GitHub, settingsClone.GithubAuth)
}

func mergeClientFromProvider(dst *Client, src settings.AuthProviderConfig) {
	if dst == nil {
		return
	}
	dst.ClientID = strings.TrimSpace(src.ClientId)
	dst.ClientSecret = strings.TrimSpace(src.ClientSecret)
	dst.Enabled = src.Enabled && dst.ClientID != "" && dst.ClientSecret != ""
}

func persistPocketBaseProviders(app *pocketbase.PocketBase, cfg Config) error {
	form := forms.NewSettingsUpsert(app)
	applyClientToProvider(&form.GoogleAuth, cfg.Google)
	applyClientToProvider(&form.GithubAuth, cfg.GitHub)
	return form.Submit()
}

func applyClientToProvider(dst *settings.AuthProviderConfig, src Client) {
	if dst == nil {
		return
	}
	dst.ClientId = strings.TrimSpace(src.ClientID)
	dst.ClientSecret = strings.TrimSpace(src.ClientSecret)
	dst.Enabled = src.Enabled && dst.ClientId != "" && dst.ClientSecret != ""
}
