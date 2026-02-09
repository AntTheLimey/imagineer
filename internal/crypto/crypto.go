/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package crypto provides AES-256-GCM encryption for sensitive data at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// KeySize is the required length in bytes for an AES-256 key.
const KeySize = 32

// Encryptor provides AES-256-GCM encryption and decryption using a
// single symmetric key. It is safe for concurrent use because the
// underlying cipher.AEAD is stateless.
type Encryptor struct {
	gcm cipher.AEAD
}

// NewEncryptor creates an Encryptor from a 32-byte AES-256 key.
// It returns an error if the key length is not exactly 32 bytes.
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("crypto: key must be exactly %d bytes, got %d", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create GCM: %w", err)
	}

	return &Encryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce and
// returns the result as a standard base64-encoded string. The encoded
// value contains the nonce prepended to the ciphertext.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	// Seal appends the ciphertext (with GCM auth tag) to the nonce slice,
	// producing nonce || ciphertext || tag.
	sealed := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return "enc:" + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decodes a base64-encoded ciphertext produced by Encrypt and
// returns the original plaintext. It returns an error if the input is
// not valid base64, is too short to contain a nonce, or fails GCM
// authentication (e.g. the data was tampered with or a different key
// was used).
func (e *Encryptor) Decrypt(encoded string) (string, error) {
	encoded = strings.TrimPrefix(encoded, "enc:")
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto: invalid base64 input: %w", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("crypto: ciphertext too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decryption failed: %w", err)
	}

	return string(plaintext), nil
}
