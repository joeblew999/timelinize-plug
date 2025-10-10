//go:build timelinize

package auth

import "context"

// AuthApple scaffolding — will require TeamID, ServiceID (ClientID), KeyID, and .p8 path.
// This is a placeholder that returns nil once wired.
func AuthApple(ctx context.Context, store interface{}) error {
	// TODO: implement SIWA: JWT client secret, code exchange; persist token via store
	return nil
}
