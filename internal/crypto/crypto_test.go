/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

// testKey returns a deterministic 32-byte key suitable for tests.
func testKey() []byte {
	return []byte("01234567890123456789012345678901") // exactly 32 bytes
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	original := "sk-live-abc123-secret-api-key"

	ciphertext, err := enc.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, original)
	}
}

func TestEncryptProducesUniqueNonces(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	plaintext := "same-plaintext-value"

	ct1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}

	ct2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if ct1 == ct2 {
		t.Error("two encryptions of the same plaintext produced identical ciphertext; nonces should differ")
	}
}

func TestNewEncryptorRejectsWrongKeyLength(t *testing.T) {
	shortKey := make([]byte, 16)
	_, err := NewEncryptor(shortKey)
	if err == nil {
		t.Fatal("expected error for 16-byte key, got nil")
	}
}

func TestDecryptDetectsTamperedCiphertext(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	ciphertext, err := enc.Encrypt("sensitive-data")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Strip the "enc:" prefix so we can decode the base64 payload,
	// tamper with a byte, re-encode, and restore the prefix.
	b64 := strings.TrimPrefix(ciphertext, "enc:")
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}

	mid := len(raw) / 2
	raw[mid] ^= 0xFF // flip all bits in one byte

	tampered := "enc:" + base64.StdEncoding.EncodeToString(raw)

	_, err = enc.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected decryption error for tampered ciphertext, got nil")
	}
}

func TestDecryptRejectsInvalidBase64(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	_, err = enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 input, got nil")
	}
}

func TestEncryptOutputHasPrefix(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	ciphertext, err := enc.Encrypt("some-api-key")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if !strings.HasPrefix(ciphertext, "enc:") {
		t.Errorf("encrypted output should start with \"enc:\" prefix, got %q", ciphertext)
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt of empty string failed: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt of empty string ciphertext failed: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}
