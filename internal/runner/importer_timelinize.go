//go:build timelinize

package runner

import (
	"context"
		"github.com/timelinize/timelinize/tlzapp"
)

// Import uses timelinize in-process (requires -tags timelinize).
func Import(ctx context.Context, source, from, out, tenant string) error {
	app, err := tlzapp.Init(ctx, nil, nil)
	if err != nil {
		return err
	}
	args := []string{
		"import",
		"--source", source,
		"--input", from,
		"--output", out,
	}
	return app.RunCommand(ctx, args)
}
