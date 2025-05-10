package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type CipherManager struct {
	secretKey []byte
}

func NewCipherManager(key string) (*CipherManager, error) {
	secretKey, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return &CipherManager{secretKey: secretKey}, nil
}

func (cm *CipherManager) Encrypt(plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(cm.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (cm *CipherManager) Decrypt(ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(cm.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
