package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
)

// SSE bridge that subscribes to NATS and streams lines to the client.
func sseRoute(r *chi.Mux, nc *nats.Conn) {
	r.Get("/events/{topic}", func(w http.ResponseWriter, req *http.Request) {
		topic := chi.URLParam(req, "topic") // import|auth|status
		if topic == "" { topic = "status" }
		tenant := req.URL.Query().Get("tenant")
		if tenant == "" { tenant = "local" }

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		subj := fmt.Sprintf("tplug.%s.%s.>", topic, tenant)
		sub, err := nc.SubscribeSync(subj)
		if err != nil {
			log.Println("sse subscribe error:", err)
			http.Error(w, "subscribe failed", http.StatusInternalServerError)
			return
		}
		defer sub.Unsubscribe()

		// initial ping
		_, _ = w.Write([]byte("event: ping\n data: ok\n\n"))
		flusher.Flush()

		for {
			msg, err := sub.NextMsg(30 * time.Second)
			if err != nil {
				_, _ = w.Write([]byte("event: ping\n data: wait\n\n"))
				flusher.Flush()
				continue
			}
			_, _ = w.Write([]byte("data: \n"))
			_, _ = w.Write(msg.Data)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	})
}

func middlewareBase(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}
