package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
	datastar "github.com/starfederation/datastar-go/datastar"
)

var allowedUITopics = map[string]struct{}{
	"auth":   {},
	"import": {},
	"status": {},
}

// SSE bridge that subscribes to NATS and streams lines to the client.
func sseRoute(r *chi.Mux, nc *nats.Conn, defaultTenant string) {
	r.Get("/events/{topic}", func(w http.ResponseWriter, req *http.Request) {
		topic := chi.URLParam(req, "topic") // import|auth|status
		if topic == "" {
			topic = "status"
		}
		tenant := req.URL.Query().Get("tenant")
		if tenant == "" {
			tenant = defaultTenant
		}

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
		_, _ = w.Write([]byte("event: ping\ndata: ok\n\n"))
		flusher.Flush()

		for {
			msg, err := sub.NextMsg(30 * time.Second)
			if err != nil {
				_, _ = w.Write([]byte("event: ping\ndata: wait\n\n"))
				flusher.Flush()
				continue
			}
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(msg.Data)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	})
}

func datastarRoute(r *chi.Mux, opt Options) {
	if opt.NATSConn == nil && opt.PBApp == nil {
		return
	}

	r.Get("/ui/datastar/{topic}", func(w http.ResponseWriter, req *http.Request) {
		topic := chi.URLParam(req, "topic")
		if topic == "" {
			topic = "status"
		}
		if _, ok := allowedUITopics[topic]; !ok {
			http.Error(w, "invalid topic", http.StatusBadRequest)
			return
		}

		tenant := req.URL.Query().Get("tenant")
		if tenant == "" {
			tenant = opt.Tenant
		}

		ctx := req.Context()
		sse := datastar.NewSSE(w, req, datastar.WithContext(ctx))
		statusTarget := fmt.Sprintf("%s-status-pill", topic)
		feedTarget := fmt.Sprintf("%s-feed", topic)
		emptyTarget := fmt.Sprintf("%s-empty", topic)

		var (
			natsStream <-chan string
			closeNATS  func()
			err        error
		)
		if opt.NATSConn != nil {
			natsStream, closeNATS, err = startNATSStream(ctx, opt.NATSConn, topic, tenant)
			if err != nil {
				log.Printf("datastar nats stream: %v", err)
				http.Error(w, "subscribe failed", http.StatusInternalServerError)
				return
			}
		}
		if closeNATS != nil {
			defer closeNATS()
		}

		var (
			pbStream <-chan string
			closePB  func()
		)
		if opt.PBApp != nil {
			pbTopics := pocketBaseTopicsFor(topic)
			if len(pbTopics) > 0 {
				var err error
				pbStream, closePB, err = startPocketBaseStream(ctx, opt.PBApp, pbTopics)
				if err != nil {
					log.Printf("datastar pocketbase stream: %v", err)
					http.Error(w, "realtime unavailable", http.StatusInternalServerError)
					return
				}
			}
		}
		if closePB != nil {
			defer closePB()
		}

		if natsStream == nil && pbStream == nil {
			http.Error(w, "no realtime sources available", http.StatusServiceUnavailable)
			return
		}

		_ = sse.PatchElements(
			`<span class="status status-live">live</span>`,
			datastar.WithSelectorID(statusTarget),
			datastar.WithModeReplace(),
		)

		removedPlaceholder := false
		activeStreams := 0
		if natsStream != nil {
			activeStreams++
		}
		if pbStream != nil {
			activeStreams++
		}

		for activeStreams > 0 {
			select {
			case <-ctx.Done():
				activeStreams = 0
			case payload, ok := <-natsStream:
				if !ok {
					natsStream = nil
					activeStreams--
					continue
				}
				if payload == "" {
					continue
				}
				if !removedPlaceholder {
					if err := sse.RemoveElementByID(emptyTarget); err == nil {
						removedPlaceholder = true
					}
				}
				if err := appendEventLine(sse, feedTarget, payload); err != nil {
					_ = sse.ConsoleError(err)
				}
			case payload, ok := <-pbStream:
				if !ok {
					pbStream = nil
					activeStreams--
					continue
				}
				if payload == "" {
					continue
				}
				if !removedPlaceholder {
					if err := sse.RemoveElementByID(emptyTarget); err == nil {
						removedPlaceholder = true
					}
				}
				if err := appendEventLine(sse, feedTarget, payload); err != nil {
					_ = sse.ConsoleError(err)
				}
			}
		}

		_ = sse.PatchElements(
			`<span class="status status-idle">idle</span>`,
			datastar.WithSelectorID(statusTarget),
			datastar.WithModeReplace(),
		)
	})
}

func appendEventLine(sse *datastar.ServerSentEventGenerator, feedTarget, payload string) error {
	// Try to format as structured progress event, fallback to simple formatting
	line := formatProgressEventHTML(payload)
	return sse.PatchElements(
		line,
		datastar.WithSelectorID(feedTarget),
		datastar.WithModeAppend(),
	)
}

func startNATSStream(ctx context.Context, nc *nats.Conn, topic, tenant string) (<-chan string, func(), error) {
	if nc == nil {
		return nil, nil, nil
	}
	subject := fmt.Sprintf("tplug.%s.%s.>", topic, tenant)
	sub, err := nc.SubscribeSync(subject)
	if err != nil {
		return nil, nil, err
	}

	stream := make(chan string, 64)

	go func() {
		defer close(stream)
		for {
			msg, err := sub.NextMsgWithContext(ctx)
			if err != nil {
				if err != context.Canceled && err != context.DeadlineExceeded &&
					err != nats.ErrConnectionClosed && err != nats.ErrNoResponders {
					log.Printf("nats stream error (%s): %v", subject, err)
				}
				return
			}
			payload := strings.TrimSpace(string(msg.Data))
			if payload == "" {
				continue
			}
			select {
			case stream <- payload:
			case <-ctx.Done():
				return
			}
		}
	}()

	cleanup := func() {
		if err := sub.Unsubscribe(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			log.Printf("nats unsubscribe %s: %v", subject, err)
		}
	}

	return stream, cleanup, nil
}

func startPocketBaseStream(ctx context.Context, app *pocketbase.PocketBase, topics []string) (<-chan string, func(), error) {
	if app == nil || len(topics) == 0 {
		return nil, nil, nil
	}

	client := subscriptions.NewDefaultClient()
	client.Subscribe(topics...)

	app.SubscriptionsBroker().Register(client)

	stream := make(chan string, 64)

	go func() {
		defer close(stream)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-client.Channel():
				if !ok {
					return
				}
				payload := strings.TrimSpace(string(msg.Data))
				if payload == "" {
					continue
				}
				line := fmt.Sprintf("[pb:%s] %s", msg.Name, payload)
				select {
				case stream <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	cleanup := func() {
		app.SubscriptionsBroker().Unregister(client.Id())
	}

	return stream, cleanup, nil
}

func pocketBaseTopicsFor(topic string) []string {
	switch topic {
	case "auth":
		return []string{"tokens"}
	case "status":
		return []string{"tokens", "kv", "secrets"}
	default:
		return nil
	}
}

func middlewareBase(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}
