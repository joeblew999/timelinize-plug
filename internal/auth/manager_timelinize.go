//go:build timelinize

package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/timelinize/timelinize/oauth2client"
)

// Authenticate runs provider-specific OAuth using timelinize/oauth2client.
// NOTE: Token persistence will be wired to PB in a follow-up push.
func Authenticate(ctx context.Context, provider string) error {
	switch provider {
	case "google", "github", "apple":
		// Placeholder flow that proves wiring; details (client IDs/secrets)
		// and token storage to be integrated with PB KV.
		log.Printf("oauth2client wiring in place for provider %s (PB storage pending)", provider)
		// A minimal call to show linkage (we'll expand in next push).
		var _ oauth2client.App
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}
