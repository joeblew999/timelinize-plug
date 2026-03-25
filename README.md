# timelinize-plug

Timelinize Plug is a self-contained CLI + server that embeds NATS (JetStream), PocketBase, and the Timelinize data pipelines into a single binary. It powers local-first timelines while streaming realtime status updates to a Datastar UI.



https://timelinize.com
https://timelinize.com/docs/
https://github.com/timelinize/timelinize

By adding Pocketbasse, we allow using to have a way to extract their data and save it in a reusable way inside Pocketbase, with authentication.

By adding NATS, we can make the system reactive.

By adding Datastar, we make the system GUI real time.

The CLI can also do some of the things the Web GUI does.

---

## Features

The web UI should guide users through OAuth setup, not require manual environment variables. 

- Embedded NATS with JetStream for import/auth/status subjects per tenant.

- Embedded PocketBase (pure Go, no CGO) for KV, secrets, and token storage.

- Datastar-driven UI panels that consume both NATS and PocketBase realtime feeds.
- PocketBase admin login + provider management UI built with Datastar (bootstrap, sign-in, toggles).
- Built-in OAuth credentials wizard (Google, GitHub) persisted to PocketBase.

- CLI commands for authentication, imports, sync, and status reporting.

- Structured realtime bridge (`tplug.db.<tenant>.<collection>.<action>`) plus PocketBase subscription fan-out.

---

## Architecture

- The living reference is [`ARCHITECTURE.md`](ARCHITECTURE.md) in the docs folder.

- Any interface or architectural change must update that document and append to the change log.

---

## Requirements
- Go 1.25+
- No additional services: NATS, PocketBase, and the web UI run in-process.

---

## Getting Started
```sh
# install dependencies
go mod tidy

# build
go build ./cmd/tplug

# or run directly
go run ./cmd/tplug serve
```

Optional flags for `serve`:
- `--memory` – in-memory JetStream storage (handy in tests).
- `--offline` – keep the local NATS instance from bridging upstream.

Set `TPLUG_TENANT_ID` to override the default `local` tenant.

http://127.0.0.1:12002

---

## CLI Commands
- `tplug serve [--memory] [--offline]` — start NATS, PocketBase, and the HTTP UI.
- `tplug auth --provider google|github|apple` — start OAuth flows and persist encrypted tokens.
- `tplug import --source <name> --from <path|url> --out <dir> [--tenant <id>]` — execute an import runner.
- `tplug status` — print build/version info.
- `tplug sync` — placeholder for future sync workflow (currently errors with “not implemented”).
- `tplug self-update` — placeholder for future self-update flow.
- `tplug config set|get` — manipulate PocketBase KV entries.

---

## Realtime Streams
- Native SSE endpoints: `/events/{topic}` with `topic ∈ {auth, import, status}`.
- Datastar UI endpoint: `/ui/datastar/{topic}` merges NATS JetStream subjects with PocketBase subscription broker events.
- NATS subjects follow `tplug.<topic>.<tenant>.*` and `tplug.db.<tenant>.<collection>.<action>`.
- PocketBase subscriptions receive mirrored events for `tokens`, `kv`, and `secrets`.

---

## Development
```sh
# format
gofmt -w ./cmd ./internal

# lint / vet as needed
go vet ./...

# tests (currently minimal)
go test ./...
```

---

## License

Apache 2.0 © Timelinize & contributors

