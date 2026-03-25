# Use Case: Real-Time Import Progress Tracking

This document describes the end-to-end implementation of real-time import progress tracking in timelinize-plug.

## Overview

The import progress tracking use case demonstrates the complete data flow from CLI import command → NATS JetStream → SSE → Datastar UI with live updates.

## Architecture Flow

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│              │         │              │         │              │
│  CLI Import  │────────▶│ NATS/JS      │────────▶│  SSE Stream  │
│  Command     │  emit   │ tplug.import │  sub    │  /events/    │
│              │  events │ .{tenant}.*  │         │  import      │
└──────────────┘         └──────────────┘         └──────────────┘
                                │                         │
                                │                         │
                                ▼                         ▼
                         ┌──────────────┐         ┌──────────────┐
                         │              │         │              │
                         │  JetStream   │         │  Datastar    │
                         │  Persistence │         │  UI Panel    │
                         │              │         │  Live Update │
                         └──────────────┘         └──────────────┘
```

## Components

### 1. Import Runner with Progress Events
**File**: `internal/runner/importer_timelinize.go`

The import runner emits structured JSON progress events:

```go
type ProgressEvent struct {
    Type      string    `json:"type"`      // progress, error, done
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`    // google_photos, github, etc
    Progress  *Progress `json:"progress,omitempty"`
    Error     string    `json:"error,omitempty"`
    Metadata  Metadata  `json:"metadata,omitempty"`
}
```

**Event Types**:
- `progress` - Import is running (emitted every 2 seconds)
- `done` - Import completed successfully
- `error` - Import failed

**NATS Subject Pattern**: `tplug.import.{tenant}.{type}`

Example subjects:
- `tplug.import.local.progress`
- `tplug.import.local.done`
- `tplug.import.local.error`

### 2. NATS Integration
**File**: `internal/nats/nats.go`

- Embedded NATS server with JetStream enabled
- Listens on `localhost:4222` (client) and `:8222` (monitoring)
- File-based persistence in `.data/nats/jetstream` (or in-memory with `--memory`)

### 3. SSE Endpoints
**Files**:
- `internal/server/sse.go` - SSE streaming logic
- `internal/server/sse_formatter.go` - Event formatting

**Endpoints**:
- `/events/import?tenant={tenant}` - Raw SSE stream (JSON events)
- `/ui/datastar/import?tenant={tenant}` - Datastar-enhanced stream (HTML patches)

### 4. Datastar UI
**File**: `internal/server/templates/index.html`

Beautiful real-time UI featuring:
- Live status indicator (idle → live)
- Progress bars and percentage
- Structured event display
- Error highlighting
- Metadata (duration, file paths, etc.)
- Auto-scroll feed
- Dark theme

## Usage

### Step 1: Start the Server

```bash
# Terminal 1: Start the server
go run ./cmd/tplug serve

# Or with in-memory NATS (for testing)
go run ./cmd/tplug serve --memory
```

The server starts:
- HTTP UI: http://127.0.0.1:12002
- NATS client: nats://127.0.0.1:4222
- NATS monitoring: http://127.0.0.1:8222

### Step 2: Open the Web UI

Open your browser to:
```
http://127.0.0.1:12002
```

You'll see three panels:
- **Import Progress** (left)
- **Auth Events** (middle)
- **System Status** (right)

### Step 3: Run an Import (requires -tags timelinize)

```bash
# Terminal 2: Run an import
# Note: This requires building with -tags timelinize and having timelinize deps
go run -tags timelinize ./cmd/tplug import \
  --source google_photos \
  --from /path/to/photos \
  --out ./_timeline
```

Without timelinize tags, you'll get:
```
Error: timelinize integration not enabled; build with -tags timelinize
```

### Step 4: Watch Live Updates

As the import runs, you'll see in the UI:

1. **Starting** event:
   ```
   Type: progress
   Message: Starting import from google_photos
   Source: google_photos
   Input: /path/to/photos
   Output: ./_timeline
   ```

2. **Progress** events (every 2 seconds):
   ```
   Type: progress
   Message: Processing items from google_photos
   Progress: 25 items processed
   Current: item-25
   ```

3. **Completion** event:
   ```
   Type: done
   Message: Import completed successfully from google_photos
   Progress: 100% • 150 items processed
   Duration: 2m34s
   ```

## Event Examples

### Progress Event
```json
{
  "type": "progress",
  "message": "Processing items from google_photos",
  "timestamp": "2025-10-14T12:30:45Z",
  "source": "google_photos",
  "progress": {
    "items_processed": 47,
    "current_item": "item-47"
  }
}
```

### Completion Event
```json
{
  "type": "done",
  "message": "Import completed successfully from google_photos",
  "timestamp": "2025-10-14T12:35:22Z",
  "source": "google_photos",
  "progress": {
    "items_processed": 150,
    "percentage": 100
  },
  "metadata": {
    "start_time": "2025-10-14T12:30:00Z",
    "end_time": "2025-10-14T12:35:22Z",
    "duration": "5m22s",
    "input_path": "/path/to/photos",
    "output_path": "./_timeline"
  }
}
```

### Error Event
```json
{
  "type": "error",
  "message": "Import failed: authentication required",
  "timestamp": "2025-10-14T12:31:15Z",
  "source": "google_photos",
  "error": "oauth token not found or expired",
  "metadata": {
    "start_time": "2025-10-14T12:30:00Z",
    "end_time": "2025-10-14T12:31:15Z",
    "duration": "1m15s"
  }
}
```

## Testing Without Timelinize

You can test the UI and event flow without timelinize by using the raw NATS client:

```bash
# Install nats CLI
go install github.com/nats-io/natscli/nats@latest

# Publish test events
nats pub tplug.import.local.progress '{
  "type": "progress",
  "message": "Test import starting",
  "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "source": "test_source",
  "progress": {"items_processed": 10}
}'

nats pub tplug.import.local.done '{
  "type": "done",
  "message": "Test import completed",
  "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "source": "test_source",
  "progress": {"items_processed": 100, "percentage": 100},
  "metadata": {"duration": "30s"}
}'
```

## Monitoring

### NATS Monitoring Dashboard
Visit http://127.0.0.1:8222 to see:
- Server info
- Connections
- Subscriptions
- JetStream streams

### Raw SSE Stream
Visit http://127.0.0.1:12002/events/import?tenant=local to see raw event stream.

### Browser DevTools
Open browser DevTools → Network → `datastar` to see SSE messages.

## Multi-Tenant Support

Import events are tenant-isolated:

```bash
# Import for tenant "user123"
go run -tags timelinize ./cmd/tplug import \
  --source google_photos \
  --from /path \
  --tenant user123

# Watch in UI: http://127.0.0.1:12002?tenant=user123
# Or via SSE: http://127.0.0.1:12002/events/import?tenant=user123
```

## Future Enhancements

1. **Real Progress Hooks**: Replace simulated progress with actual timelinize progress callbacks
2. **Progress Percentage**: Calculate accurate completion percentage based on total items
3. **Throughput Metrics**: Items/second, bytes/second
4. **Pause/Resume**: Add controls to pause/resume imports
5. **Detailed Item View**: Click on items to see details
6. **Export Logs**: Download import logs as JSON/CSV
7. **Notifications**: Browser notifications on completion/errors

## Troubleshooting

### No events showing in UI
- Check that NATS is running: `curl http://127.0.0.1:8222/varz`
- Check browser DevTools console for errors
- Verify SSE connection: Network tab should show `/ui/datastar/import` as "pending"

### Import fails immediately
- Check if OAuth token exists: `tplug auth --provider google`
- Verify input path exists and is readable
- Check logs in terminal running `tplug import`

### UI shows "idle" but import is running
- Check tenant ID matches between import and UI
- Verify NATS subject: `nats sub "tplug.import.>"`

## Related Files

- [internal/runner/importer_timelinize.go](../internal/runner/importer_timelinize.go) - Import logic
- [internal/runner/progress.go](../internal/runner/progress.go) - NATS emit helper
- [internal/nats/nats.go](../internal/nats/nats.go) - Embedded NATS
- [internal/server/sse.go](../internal/server/sse.go) - SSE streaming
- [internal/server/sse_formatter.go](../internal/server/sse_formatter.go) - HTML formatting
- [internal/server/templates/index.html](../internal/server/templates/index.html) - UI
- [cmd/tplug/import.go](../../cmd/tplug/import.go) - CLI command
- [cmd/tplug/serve.go](../../cmd/tplug/serve.go) - Server command
