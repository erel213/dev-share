package crypto

import (
	"crypto/rand"
	"testing"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := "my-secret-value-123"
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDifferentCiphertextsPerCall(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := "same-plaintext"
	ct1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt 1 failed: %v", err)
	}
	ct2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt 2 failed: %v", err)
	}

	if ct1 == ct2 {
		t.Error("expected different ciphertexts for same plaintext (random nonce)")
	}

	// Both should still decrypt correctly
	d1, _ := enc.Decrypt(ct1)
	d2, _ := enc.Decrypt(ct2)
	if d1 != plaintext || d2 != plaintext {
		t.Error("both ciphertexts should decrypt to the same plaintext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := generateTestKey(t)
	key2 := generateTestKey(t)

	enc1, _ := NewAESEncryptor(key1)
	enc2, _ := NewAESEncryptor(key2)

	ciphertext, _ := enc1.Encrypt("secret")
	_, err := enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	key := generateTestKey(t)
	enc, _ := NewAESEncryptor(key)

	_, err := enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for corrupted base64")
	}

	_, err = enc.Decrypt("dGVzdA==") // valid base64, but not encrypted
	if err == nil {
		t.Error("expected error for non-encrypted data")
	}
}

func TestInvalidKeySize(t *testing.T) {
	_, err := NewAESEncryptor([]byte("too-short"))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestEncryptEmptyString(t *testing.T) {
	key := generateTestKey(t)
	enc, _ := NewAESEncryptor(key)

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("encrypt empty string failed: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}
