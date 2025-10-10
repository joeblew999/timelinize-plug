package runner

// Default (no-tag) stub for timelinize integration.
// Build with -tags timelinize to enable direct import.

import (
	"context"
	"errors"
)

func Import(ctx context.Context, source, from, out, tenant string) error {
	return errors.New("timelinize integration not enabled; build with -tags timelinize")
}
