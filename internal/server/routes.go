package server

import (
	"context"
	"embed"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html
var templatesFS embed.FS

func attachRoutes(r *chi.Mux, opt Options) {
	t := template.Must(template.ParseFS(templatesFS, "templates/*.html"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Tenant string
		}{
			Tenant: opt.Tenant,
		}
		_ = t.ExecuteTemplate(w, "index.html", data)
	})

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Tenant string
		}{
			Tenant: opt.Tenant,
		}
		_ = t.ExecuteTemplate(w, "index.html", data) // reuse for now
	})
}

func Start(ctx context.Context, opt Options) error {
	r := chi.NewRouter()
	middlewareBase(r)
	attachRoutes(r, opt)
	registerOAuthRoutes(r, opt.PBApp)
	registerPocketBaseAuthRoutes(r, opt)
	sseRoute(r, opt.NATSConn, opt.Tenant)
	datastarRoute(r, opt)

	s := &http.Server{Addr: opt.Addr, Handler: r}
	go func() { <-ctx.Done(); _ = s.Shutdown(context.Background()) }()
	return s.ListenAndServe()
}
