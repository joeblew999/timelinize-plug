module github.com/joeblew999/timelinize-plug

go 1.23

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/nats-io/nats-server/v2 v2.10.11
	github.com/nats-io/nats.go v1.36.0
	github.com/pocketbase/pocketbase v0.22.14
	github.com/spf13/cobra v1.8.1
	golang.org/x/oauth2 v0.24.0
)

// Optional: keep modernc/sqlite (PocketBase default) to avoid CGO.
// No cgo needed.
