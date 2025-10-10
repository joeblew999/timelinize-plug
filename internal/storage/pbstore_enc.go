package storage

import (
	"context"
	"errors"
	"github.com/pocketbase/pocketbase"
)

// PBTokenStore now encrypts at rest using AES-GCM key from PB secrets.
type PBTokenStoreEnc struct {
	App *pocketbase.PocketBase
}

func (s *PBTokenStoreEnc) Save(ctx context.Context, provider string, data []byte) error {
	if s.App == nil { return errors.New("PB app is nil") }
	key, err := GetOrCreateAESKey(s.App)
	if err != nil { return err }
	a, err := NewAESGCM(key)
	if err != nil { return err }
	ct, err := a.Seal(data)
	if err != nil { return err }
	// reuse existing PBTokenStore impl for persistence (unencrypted payload replaced by ciphertext base64)
	base := &PBTokenStore{App: s.App}
	return base.Save(ctx, provider, ct)
}

func (s *PBTokenStoreEnc) Load(ctx context.Context, provider string) ([]byte, error) {
	if s.App == nil { return nil, errors.New("PB app is nil") }
	key, err := GetOrCreateAESKey(s.App)
	if err != nil { return nil, err }
	base := &PBTokenStore{App: s.App}
	ct, err := base.Load(ctx, provider)
	if err != nil { return nil, err }
	a, err := NewAESGCM(key)
	if err != nil { return nil, err }
	pt, err := a.Open(ct)
	if err != nil { return nil, err }
	return pt, nil
}
