package auth

// Default (no build tag) stub that compiles without timelinize.
// Real oauth via timelinize/oauth2client lives in manager_timelinize.go

import (
	"context"
	"errors"
)

func Authenticate(ctx context.Context, provider string) error {
	return errors.New("oauth not enabled; build with -tags timelinize")
}
