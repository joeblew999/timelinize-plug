package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

// AESGCM wraps a key for encrypt/decrypt.
type AESGCM struct { key []byte }

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 { return nil, errors.New("key must be 32 bytes") }
	return &AESGCM{key: key}, nil
}

func (a *AESGCM) Seal(plain []byte) ([]byte, error) {
	blk, err := aes.NewCipher(a.key)
	if err != nil { return nil, err }
	gcm, err := cipher.NewGCM(blk)
	if err != nil { return nil, err }
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil { return nil, err }
	ct := gcm.Seal(nonce, nonce, plain, nil)
	return ct, nil
}

func (a *AESGCM) Open(ct []byte) ([]byte, error) {
	blk, err := aes.NewCipher(a.key)
	if err != nil { return nil, err }
	gcm, err := cipher.NewGCM(blk)
	if err != nil { return nil, err }
	ns := gcm.NonceSize()
	if len(ct) < ns { return nil, errors.New("ciphertext too short") }

	nonce := ct[:ns]
	data := ct[ns:]
	return gcm.Open(nil, nonce, data, nil)
}
