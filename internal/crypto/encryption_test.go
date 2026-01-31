package crypto

import (
"os"
"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := NewEncryptionKey("test-encryption-key-12345")
	
	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "my-api-key-12345"},
		{"empty string", ""},
		{"long text", "this is a very long API key that should still be encrypted properly without any issues at all"},
		{"special chars", "key!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := key.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}
			
			if tt.plaintext == "" && ciphertext != nil {
				t.Errorf("Expected nil ciphertext for empty plaintext")
			}
			
			// Decrypt
			decrypted, err := key.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}
			
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text doesn't match original. Got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptDecryptBase64(t *testing.T) {
	key := NewEncryptionKey("test-encryption-key-67890")
	plaintext := "my-secret-api-key"
	
	// Encrypt to base64
	encrypted, err := key.EncryptToBase64(plaintext)
	if err != nil {
		t.Fatalf("EncryptToBase64 failed: %v", err)
	}
	
	if encrypted == "" {
		t.Fatal("Expected non-empty encrypted string")
	}
	
	// Decrypt from base64
	decrypted, err := key.DecryptFromBase64(encrypted)
	if err != nil {
		t.Fatalf("DecryptFromBase64 failed: %v", err)
	}
	
	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match. Got %q, want %q", decrypted, plaintext)
	}
}

func TestDifferentKeys(t *testing.T) {
	key1 := NewEncryptionKey("key1")
	key2 := NewEncryptionKey("key2")
	
	plaintext := "secret-data"
	
	// Encrypt with key1
	ciphertext, err := key1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	
	// Try to decrypt with key2 (should fail)
	_, err = key2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Expected decryption with wrong key to fail, but it succeeded")
	}
}

func TestGetEncryptionKeyFromEnv(t *testing.T) {
	// Test with key set
	os.Setenv("ENCRYPTION_KEY", "test-key")
	defer os.Unsetenv("ENCRYPTION_KEY")
	
	key, err := GetEncryptionKeyFromEnv()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if key == nil {
		t.Fatal("Expected key to be non-nil")
	}
	
	// Test without key set
	os.Unsetenv("ENCRYPTION_KEY")
	_, err = GetEncryptionKeyFromEnv()
	if err == nil {
		t.Error("Expected error when ENCRYPTION_KEY not set")
	}
}

func TestEncryptSameTextDifferentCiphertexts(t *testing.T) {
	key := NewEncryptionKey("test-key")
	plaintext := "same-text"
	
	// Encrypt twice
	cipher1, err := key.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("First encrypt failed: %v", err)
	}
	
	cipher2, err := key.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second encrypt failed: %v", err)
	}
	
	// Ciphertexts should be different (due to random nonce)
	if string(cipher1) == string(cipher2) {
		t.Error("Expected different ciphertexts for same plaintext (nonce should randomize)")
	}
	
	// But both should decrypt to same plaintext
	decrypted1, _ := key.Decrypt(cipher1)
	decrypted2, _ := key.Decrypt(cipher2)
	
	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to same plaintext")
	}
}
