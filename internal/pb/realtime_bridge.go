package pb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

// RealtimeBridge fans PocketBase lifecycle events out to both the
// embedded PocketBase subscription broker and JetStream/NATS subjects.
type RealtimeBridge struct {
	app       *pocketbase.PocketBase
	conn      *nats.Conn
	js        nats.JetStreamContext
	tenant    string
	whitelist map[string]struct{}
}

// NewRealtimeBridge constructs a bridge. Tenant defaults to "local".
func NewRealtimeBridge(app *pocketbase.PocketBase, conn *nats.Conn, js nats.JetStreamContext, tenant string) *RealtimeBridge {
	if strings.TrimSpace(tenant) == "" {
		tenant = "local"
	}

	return &RealtimeBridge{
		app:    app,
		conn:   conn,
		js:     js,
		tenant: tenant,
		whitelist: map[string]struct{}{
			"tokens":  {},
			"kv":      {},
			"secrets": {},
		},
	}
}

// Start wires the PocketBase hooks if both app and transport are available.
func (b *RealtimeBridge) Start() {
	if b.app == nil {
		return
	}

	dispatch := func(action string, rec *models.Record) error {
		b.forward(action, rec)
		return nil
	}

	b.app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		return dispatch("create", e.Record)
	})
	b.app.OnRecordAfterUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
		return dispatch("update", e.Record)
	})
	b.app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		return dispatch("delete", e.Record)
	})
}

func (b *RealtimeBridge) forward(action string, rec *models.Record) {
	if rec == nil {
		return
	}
	if _, ok := b.whitelist[rec.Collection().Name]; !ok {
		return
	}

	exported := sanitizeRecord(rec)
	payload := map[string]any{
		"ts":         time.Now().Unix(),
		"tenant":     b.tenant,
		"collection": rec.Collection().Name,
		"action":     action,
		"recordId":   rec.Id,
	}
	if len(exported) > 0 {
		payload["record"] = exported
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("realtime bridge marshal: %v", err)
		return
	}

	b.publishNATS(data, rec.Collection().Name, action)
	b.broadcastToBroker(action, rec, exported)
}

func (b *RealtimeBridge) publishNATS(data []byte, collection, action string) {
	if b.conn == nil && b.js == nil {
		return
	}

	// status feed for existing UI
	if b.conn != nil {
		statusSubject := fmt.Sprintf("tplug.status.%s.pb", b.tenant)
		if err := b.conn.Publish(statusSubject, data); err != nil {
			log.Printf("realtime bridge publish %s: %v", statusSubject, err)
		}
	}

	// structured JetStream subject (tenanted)
	subject := fmt.Sprintf("tplug.db.%s.%s.%s", b.tenant, collection, action)
	switch {
	case b.js != nil:
		if _, err := b.js.Publish(subject, data); err != nil {
			log.Printf("realtime bridge jetstream %s: %v", subject, err)
		}
	case b.conn != nil:
		if err := b.conn.Publish(subject, data); err != nil {
			log.Printf("realtime bridge publish %s: %v", subject, err)
		}
	}
}

func (b *RealtimeBridge) broadcastToBroker(action string, rec *models.Record, exported map[string]any) {
	if b.app == nil {
		return
	}

	body := map[string]any{
		"action": action,
	}
	if len(exported) > 0 {
		body["record"] = exported
	}

	raw, err := json.Marshal(body)
	if err != nil {
		log.Printf("realtime bridge broker marshal: %v", err)
		return
	}

	topics := []string{
		rec.Collection().Name,
		fmt.Sprintf("%s/%s", rec.Collection().Name, rec.Id),
	}

	for _, client := range b.app.SubscriptionsBroker().Clients() {
		if client == nil || client.IsDiscarded() {
			continue
		}
		sent := make(map[string]struct{})
		for _, topic := range topics {
			for name := range client.Subscriptions(topic) {
				if _, exists := sent[name]; exists {
					continue
				}
				sent[name] = struct{}{}
				msg := subscriptions.Message{
					Name: name,
					Data: raw,
				}
				go client.Send(msg)
			}
		}
	}
}

func sanitizeRecord(rec *models.Record) map[string]any {
	if rec == nil {
		return nil
	}
	out := rec.PublicExport()
	switch rec.Collection().Name {
	case "tokens", "secrets":
		delete(out, "payload")
		delete(out, "value")
	}
	return out
}
