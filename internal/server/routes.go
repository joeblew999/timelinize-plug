package server

import (
	"context"
	"embed"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

//go:embed templates/*.html
var templatesFS embed.FS

func attachRoutes(r *chi.Mux, nc *nats.Conn) {
	t := template.Must(template.ParseFS(templatesFS, "templates/*.html"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_ = t.ExecuteTemplate(w, "index.html", nil)
	})

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		_ = t.ExecuteTemplate(w, "index.html", nil) // reuse for now
	})
}

func Start(ctx context.Context, opt Options) error {
	r := chi.NewRouter()
	attachRoutes(r, opt.NATSConn)
	sseRoute(r, opt.NATSConn)

	s := &http.Server{Addr: opt.Addr, Handler: r}
	go func() { <-ctx.Done(); _ = s.Shutdown(context.Background()) }()
	return s.ListenAndServe()
}
