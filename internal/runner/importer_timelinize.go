//go:build timelinize

package runner

import (
	"context"
	"encoding/json"
	"time"
		"github.com/timelinize/timelinize/tlzapp"
)

// Import uses timelinize in-process (requires -tags timelinize) and emits NATS events.
func Import(ctx context.Context, source, from, out, tenant string) error {
	emit("import", tenant, "progress", []byte("{\"msg\":\"start\"}"))
	start := time.Now()
	app, err := tlzapp.Init(ctx, nil, nil)
	if err != nil {
		emit("import", tenant, "error", []byte("{\"msg\":\"init failed\"}"))
		return err
	}
	args := []string{
		"import",
		"--source", source,
		"--input", from,
		"--output", out,
	}
	if err := app.RunCommand(ctx, args); err != nil {
		emit("import", tenant, "error", []byte("{\"msg\":\"failed\"}"))
		return err
	}
	meta,_ := json.Marshal(map[string]any{
		"msg": "done",
		"took": time.Since(start).String(),
	})
	emit("import", tenant, "done", meta)
	return nil
}
