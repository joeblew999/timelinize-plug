package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// LocalRedirectServer helps run a temporary HTTP server to capture OAuth codes.
type LocalRedirectServer struct {
	Addr string // e.g., 127.0.0.1:8008
}

func (s *LocalRedirectServer) WaitCode(ctx context.Context, state string, timeout time.Duration) (string, error) {
	if s.Addr == "" { s.Addr = "127.0.0.1:8008" }
	ch := make(chan string, 1)
	server := &http.Server{Addr: s.Addr}
	http.HandleFunc("/oauth2-redirect", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state mismatch", 400)
			return
		}
		code := r.URL.Query().Get("code")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"message": "you can close this tab",
		})
		ch <- code
		go server.Shutdown(context.Background())
	})
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(timeout):
		return "", errors.New("oauth timeout")
	case code := <-ch:
		return code, nil
	}
}

// Exchange is a convenience wrapper around oauth2.Config.Exchange with context and timeout.
func Exchange(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return conf.Exchange(ctx2, code)
}
