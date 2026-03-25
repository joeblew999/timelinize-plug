package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/timelinize-plug/internal/oauth"
	"github.com/pocketbase/pocketbase"
)

func registerOAuthRoutes(r *chi.Mux, app *pocketbase.PocketBase) {
	if app == nil {
		return
	}

	r.Route("/api/oauth", func(r chi.Router) {
		r.Get("/clients", func(w http.ResponseWriter, req *http.Request) {
			cfg, err := oauth.Load(app)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, cfg)
		})

		r.Post("/clients", func(w http.ResponseWriter, req *http.Request) {
			cfg := oauth.Default()
			if err := json.NewDecoder(req.Body).Decode(&cfg); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			oauth.Sanitize(&cfg.Google)
			oauth.Sanitize(&cfg.GitHub)
			if err := oauth.Save(app, cfg); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, cfg)
		})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
