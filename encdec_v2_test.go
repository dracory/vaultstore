package vaultstore

import (
	"strconv"
	"strings"
	"testing"
)

// Test v2 encryption constants
func TestEncryptionVersionConstants(t *testing.T) {
	if ENCRYPTION_VERSION_V1 != "v1" {
		t.Errorf("ENCRYPTION_VERSION_V1 expected 'v1', got '%s'", ENCRYPTION_VERSION_V1)
	}
	if ENCRYPTION_VERSION_V2 != "v2" {
		t.Errorf("ENCRYPTION_VERSION_V2 expected 'v2', got '%s'", ENCRYPTION_VERSION_V2)
	}
	if ENCRYPTION_PREFIX_V1 != "v1:" {
		t.Errorf("ENCRYPTION_PREFIX_V1 expected 'v1:', got '%s'", ENCRYPTION_PREFIX_V1)
	}
	if ENCRYPTION_PREFIX_V2 != "v2:" {
		t.Errorf("ENCRYPTION_PREFIX_V2 expected 'v2:', got '%s'", ENCRYPTION_PREFIX_V2)
	}
}

// Test v2 encryption/decryption roundtrip
func Test_encodeV2_decodeV2_Roundtrip(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		password string
	}{
		{"simple", "test_value", "test_password"},
		{"empty", "", "password"},
		{"long value", createRandomBlock(10000), "password"},
		{"unicode", "Hello, 世界! 🌍", "unicode_password_日本語"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?", "complex_pass"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := encodeV2(tc.value, tc.password, nil)
			if err != nil {
				t.Fatalf("encodeV2 failed: %v", err)
			}
			if !strings.HasPrefix(encoded, ENCRYPTION_PREFIX_V2) {
				t.Fatalf("Expected v2: prefix, got: %s", encoded[:10])
			}

			decoded, err := decodeV2(encoded, tc.password, nil)
			if err != nil {
				t.Fatalf("decodeV2 failed: %v", err)
			}

			if decoded != tc.value {
				t.Fatalf("Roundtrip failed: expected %q, got %q", tc.value, decoded)
			}
		})
	}
}

// Test v2 with wrong password
func Test_decodeV2_WrongPassword(t *testing.T) {
	value := "secret data"
	password := "correct_password"
	wrongPassword := "wrong_password"

	encoded, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}
	_, err = decodeV2(encoded, wrongPassword, nil)
	if err == nil {
		t.Fatal("Expected error with wrong password, got nil")
	}
}

// Test v2 tampered ciphertext
func Test_decodeV2_TamperedCiphertext(t *testing.T) {
	value := "secret data"
	password := "password"

	encoded, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}
	// Tamper with the ciphertext by changing a character
	tampered := encoded[:len(encoded)-5] + "XXXXX"

	_, err = decodeV2(tampered, password, nil)
	if err == nil {
		t.Fatal("Expected error with tampered ciphertext, got nil")
	}
}

// Test v2 with empty password
func Test_encodeV2_decodeV2_EmptyPassword(t *testing.T) {
	value := "test data"
	password := ""

	encoded, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}
	decoded, err := decodeV2(encoded, password, nil)
	if err != nil {
		t.Fatalf("decodeV2 with empty password failed: %v", err)
	}
	if decoded != value {
		t.Fatalf("Roundtrip with empty password failed: expected %q, got %q", value, decoded)
	}
}

// Test encode() always uses v2
func Test_encode_UsesV2(t *testing.T) {
	value := "test_value"
	password := "test_password"

	encoded, err := encode(value, password, nil)
	if err != nil {
		t.Fatalf("encode() failed: %v", err)
	}
	if !strings.HasPrefix(encoded, ENCRYPTION_PREFIX_V2) {
		t.Fatalf("encode() should use v2 prefix, got: %s", encoded[:10])
	}
}

// Test decode() handles v2
func Test_decode_HandlesV2(t *testing.T) {
	value := "test_value"
	password := "test_password"

	encoded, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}
	decoded, err := decode(encoded, password, nil)
	if err != nil {
		t.Fatalf("decode() failed for v2: %v", err)
	}
	if decoded != value {
		t.Fatalf("decode() v2 failed: expected %q, got %q", value, decoded)
	}
}

// Test decode() backward compatibility with v1 (legacy)
func Test_decode_BackwardCompatibilityV1(t *testing.T) {
	// This tests that existing v1 encrypted data still works
	value := "test_value"
	password := "test_password"

	// Manually create v1 encoded data using legacy encode method
	legacyEncoded := encodeV1(value, password)

	// decode() should handle v1 data
	decoded, err := decode(legacyEncoded, password, nil)
	if err != nil {
		t.Fatalf("decode() failed for v1 legacy data: %v", err)
	}
	if decoded != value {
		t.Fatalf("decode() v1 legacy failed: expected %q, got %q", value, decoded)
	}
}

// Test decodeV2 with invalid base64
func Test_decodeV2_InvalidBase64(t *testing.T) {
	invalidEncoded := ENCRYPTION_PREFIX_V2 + "!!!not_valid_base64!!!"
	_, err := decodeV2(invalidEncoded, "password", nil)
	if err == nil {
		t.Fatal("Expected error for invalid base64, got nil")
	}
}

// Test decodeV2 with too short data
func Test_decodeV2_TooShortData(t *testing.T) {
	shortData := ENCRYPTION_PREFIX_V2 + base64Encode([]byte("short"))
	_, err := decodeV2(shortData, "password", nil)
	if err == nil {
		t.Fatal("Expected error for short data, got nil")
	}
}

// Test Argon2id key derivation produces consistent keys
func Test_deriveKeyArgon2id_Consistency(t *testing.T) {
	password := "test_password"
	salt := []byte("1234567890123456")

	key1 := deriveKeyArgon2id(password, salt, nil)
	key2 := deriveKeyArgon2id(password, salt, nil)

	if string(key1) != string(key2) {
		t.Fatal("Argon2id key derivation not deterministic for same salt")
	}

	if len(key1) != ARGON2_KEY_LENGTH {
		t.Fatalf("Expected key length %d, got %d", ARGON2_KEY_LENGTH, len(key1))
	}
}

// Test different salts produce different keys
func Test_deriveKeyArgon2id_DifferentSalts(t *testing.T) {
	password := "test_password"
	salt1 := []byte("1234567890123456")
	salt2 := []byte("6543210987654321")

	key1 := deriveKeyArgon2id(password, salt1, nil)
	key2 := deriveKeyArgon2id(password, salt2, nil)

	if string(key1) == string(key2) {
		t.Fatal("Different salts should produce different keys")
	}
}

// Test that each v2 encryption produces different output (due to random salt)
func Test_encodeV2_UniqueOutput(t *testing.T) {
	value := "same_value"
	password := "same_password"

	encoded1, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}
	encoded2, err := encodeV2(value, password, nil)
	if err != nil {
		t.Fatalf("encodeV2 failed: %v", err)
	}

	if encoded1 == encoded2 {
		t.Fatal("encodeV2 should produce unique output each time due to random salt")
	}

	// Both should decrypt to same value
	decoded1, _ := decodeV2(encoded1, password, nil)
	decoded2, _ := decodeV2(encoded2, password, nil)

	if decoded1 != value || decoded2 != value {
		t.Fatal("Both encodings should decrypt to original value")
	}
}

// encodeV1 is the legacy encoding method for testing backward compatibility
func encodeV1(value string, password string) string {
	strongPassword := strongifyPassword(password)
	v1 := base64Encode([]byte(value))
	v2 := strconv.Itoa(len(v1)) + "_" + v1
	randomBlock := createRandomBlock(calculateRequiredBlockLength(len(v2)))
	v3 := v2 + "" + randomBlock[len(v2):]
	v4 := base64Encode([]byte(v3))
	last := xorEncrypt(v4, strongPassword)
	return last
}
