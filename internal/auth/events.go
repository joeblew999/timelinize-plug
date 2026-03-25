package auth

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Emit auth lifecycle events to NATS so UI can stream via SSE.
func publishAuthEvent(nc *nats.Conn, tenant, kind string, payload []byte) error {
	if tenant == "" {
		tenant = "local"
	}
	if kind == "" {
		kind = "status"
	}
	subj := fmt.Sprintf("tplug.auth.%s.%s", tenant, kind) // e.g., tplug.auth.local.progress
	return nc.Publish(subj, payload)
}

func publishAuthPing(nc *nats.Conn, tenant string) {
	_ = publishAuthEvent(nc, tenant, "progress", []byte(fmt.Sprintf("{\"ts\":%d,\"msg\":\"auth starting\"}", time.Now().Unix())))
}
