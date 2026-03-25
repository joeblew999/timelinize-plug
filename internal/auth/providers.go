package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/joeblew999/timelinize-plug/internal/storage"
)

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

func randomState(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// AuthGoogle performs local OAuth with browser + redirect then persists token.
func AuthGoogle(ctx context.Context, store storage.TokenStore) error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirect := "http://127.0.0.1:8008/oauth2-redirect"
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/photoslibrary.readonly",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		Endpoint: google.Endpoint,
	}
	state := randomState(18)
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	openBrowser(url)
	ls := &LocalRedirectServer{Addr: "127.0.0.1:8008"}
	code, err := ls.WaitCode(ctx, state, 3*time.Minute)
	if err != nil {
		return err
	}
	tok, err := Exchange(ctx, conf, code)
	if err != nil {
		return err
	}
	return PersistRawToken(ctx, store, "google", tok)
}

// AuthGitHub similar to Google.
func AuthGitHub(ctx context.Context, store storage.TokenStore) error {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirect := "http://127.0.0.1:8008/oauth2-redirect"
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes:       []string{"repo"},
		Endpoint:     github.Endpoint,
	}
	state := randomState(18)
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	openBrowser(url)
	ls := &LocalRedirectServer{Addr: "127.0.0.1:8008"}
	code, err := ls.WaitCode(ctx, state, 3*time.Minute)
	if err != nil {
		return err
	}
	tok, err := Exchange(ctx, conf, code)
	if err != nil {
		return err
	}
	return PersistRawToken(ctx, store, "github", tok)
}
