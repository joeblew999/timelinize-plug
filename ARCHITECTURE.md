# timelinize-plug — Architecture (Living Doc)

> This file is authoritative. Any architectural/interface change must update this doc and append a dated entry in **Change Log**.

## TL;DR
- Single binary `tplug`: CLI (Cobra) + Server (Chi + html/template + SSE).
- Embedded **NATS (JetStream, leaf)** + embedded **PocketBase (NO CGO)** for KV/files.
- Direct import of **github.com/timelinize/timelinize** (datasources).
- Auth via **timelinize/oauth2client** (Google, GitHub; Apple later).
- Files via PB: local FS or S3/R2 (toggle in Go).
- Self-update: manual cmd + daily server prompt (HTTPS integrity for now).

## Defaults
- HTTP UI/API: `127.0.0.1:12002`
- OAuth redirect: `127.0.0.1:8008` (`/oauth2-redirect`)
- NATS client/monitor/routes: `4222 / 8222 / 6222`
- Tenant: dev=`local`; prod via `TPLUG_TENANT_ID` (fallback `local`)
- Config precedence: **flags → env → tplug.yaml → built-ins** (runs with zero flags)

## Embedded NATS (leaf)
- Always start local NATS + JetStream.
- Optional bridge upstream when `TPLUG_NATS_URL` + creds env set (`TPLUG_NATS_CREDS` or JWT/NKEY).
- JetStream store: **file** under `~/.tplug/nats/jetstream`; `--memory` for tests.
- Subjects/Streams (per tenant)
  - Streams: `tplug.import.<tenant>`, `tplug.auth.<tenant>`, `tplug.status.<tenant>`
  - Subjects: `*.progress`, `*.done`, `*.error`
- SSE topics: `/events/import`, `/events/auth`, `/events/status`

## PocketBase (NO CGO)
- Embedded PB provides SQLite, KV/Secrets, Files (local FS or S3/R2 toggle).
- Layout: local `~/Timelinize/<source>/...`; multi-tenant `tenants/<tenant_id>/timeline/...`
- Token encryption: PB-managed AES key (fallback: local key file).
- Later: PB auth for UI users (Datastar GUI).

## Timelinize & Datasources
- Import `github.com/timelinize/timelinize` directly (no subprocess).
- Initial: Google (Photos, Gmail), GitHub, Generic (JSON/CSV).
- Apple: placeholders now; add TeamID/ServiceID/KeyID/.p8 later (browser login).

## OAuth (oauth2client)
- GUI: open browser, local redirect; Headless: code-paste fallback.
- Credentials: local YAML defaults (`~/.tplug/credentials.yaml`) + PB KV overrides (`oauth2/clients`).

## CLI
- `tplug auth --provider google|github|apple`
- `tplug import --source <name> --from <path|url> --out <dir> [--tenant <id>]`
- `tplug serve [--memory] [--offline]`
- `tplug status`
- `tplug sync`
- `tplug self-update`

## HTTP
- `GET /` UI (html/template)
- `POST /api/auth/:provider`
- `POST /api/import`
- `GET /api/status`
- `GET /events/{topic}` (`import`|`auth`|`status`)
- `GET /healthz`

## Diagram
```mermaid
flowchart LR
  subgraph BIN["tplug (single binary)"]
    subgraph NATS["Embedded NATS + JetStream"]
      N1[(Per-tenant streams)]
    end
    subgraph PB["Embedded PocketBase (NO CGO)"]
      KV[(KV/Secrets)]
      FS[(Files: FS or S3/R2)]
      DB[(SQLite)]
    end
    subgraph WEB["Chi + html/template + SSE"]
      UI[UI]
      SSE[/SSE endpoints/]
      API[REST API]
    end
    subgraph TLZ["timelinize (module)"]
      DS1[Google] DS2[GitHub] DS3[Generic] DS4[Apple*]
    end
    OAUTH[oAuth2client]

    UI-->API
    API-->TLZ
    TLZ-->PB
    TLZ-->N1
    N1-->SSE
    OAUTH-->TLZ
  end
  EXT[(Upstream NATS)] --- N1
```

## Logging, Shutdown, CI
- Dev: stdout JSON; Prod: stdout + `~/.tplug/logs/*.json`
- SIGINT/SIGTERM: drain NATS, flush logs, close DB.
- CI: build/vet/test; release matrix (linux/darwin arm64/amd64, windows amd64). PB built **NO CGO**.

## Roadmap
- [ ] P1: CLI/Server, embedded NATS, PB stubs, SSE topics
- [ ] P1: timelinize import (Google, GitHub, Generic) via runner
- [ ] P1: oauth2client flows + token encryption (PB)
- [ ] P1: self-update cmd + server daily check
- [ ] P2: Apple provider flow + keys
- [ ] P2: PB auth for UI; Datastar GUI integration
- [ ] P2: S3/R2 toggle + multi-tenant finalize
- [ ] P3: checksums/signatures, metrics, tracing

## Change Log
- [2025-10-10] Initial spec committed.
