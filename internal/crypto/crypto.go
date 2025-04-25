package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type Crypto struct {
	aesGCM cipher.AEAD
}

func (c *Crypto) Encrypt(data []byte) (string, error) {
	log := logger.GetDefault()
	nonce := make([]byte, c.aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		log.Error("failed to generate nonce", "error", err)
		return "", err
	}

	ciphertext := c.aesGCM.Seal(nonce, nonce, data, nil)
	return hex.EncodeToString(ciphertext), nil
}

func (c *Crypto) Decrypt(encryptedData string) ([]byte, error) {
	log := logger.GetDefault()
	ciphertext, err := hex.DecodeString(encryptedData)
	if err != nil {
		log.Error("failed to decode ciphertext", "error", err)
		return nil, err
	}

	nonceSize := c.aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Error("ciphertext too short")
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := c.aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Error("failed to decrypt data", "error", err)
		return nil, err
	}

	return plaintext, nil
}

func NewCrypto(key []byte) (*Crypto, error) {
	if len(key) < 8 {
		return nil, fmt.Errorf("the key is too short, min length is 8")
	}
	hashed := sha256.Sum256(key) // Always a 32-byte AES-256 key

	block, err := aes.NewCipher(hashed[:])
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Crypto{aesGCM: aesGCM}, nil
}
