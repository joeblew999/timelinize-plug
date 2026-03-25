package server

import (
	"github.com/nats-io/nats.go"
	"github.com/pocketbase/pocketbase"
)

// Options encapsulates the dependencies required to start the HTTP server.
type Options struct {
	Addr        string
	OAuthCBAddr string
	JSCtx       nats.JetStreamContext
	NATSConn    *nats.Conn
	PBApp       *pocketbase.PocketBase
	Tenant      string
}
