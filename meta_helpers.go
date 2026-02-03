package vaultstore

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword creates an Argon2id hash of the password
// This uses memory-hard hashing for better resistance against GPU/ASIC attacks
func hashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, ARGON2ID_SALT_LEN)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate Argon2id hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		ARGON2ID_TIME,
		ARGON2ID_MEMORY,
		ARGON2ID_THREADS,
		ARGON2ID_KEY_LEN,
	)

	// Encode as base64 for storage
	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		ARGON2ID_MEMORY,
		ARGON2ID_TIME,
		ARGON2ID_THREADS,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encodedHash, nil
}

// verifyPassword verifies a password against a hash
// Supports both Argon2id (new) and bcrypt (legacy) hashes for backward compatibility
func verifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	// Check if it's an Argon2id hash
	if strings.HasPrefix(hash, "$argon2id$") {
		return verifyArgon2id(password, hash)
	}

	// Check if it's a bcrypt hash (starts with $2a$, $2b$, or $2y$)
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		return verifyBcrypt(password, hash)
	}

	// Unknown hash format
	return false
}

// verifyArgon2id verifies a password against an Argon2id hash
func verifyArgon2id(password, encodedHash string) bool {
	// Parse the encoded hash
	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}

	// Parse parameters
	var memory, time, threads int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	// Decode expected hash
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Compute hash with same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(time),
		uint32(memory),
		uint8(threads),
		uint32(len(expectedHash)),
	)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1
}

// verifyBcrypt verifies a password against a bcrypt hash (legacy)
func verifyBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generatePasswordID generates a unique password identity ID
func generatePasswordID() string {
	return PASSWORD_ID_PREFIX + uuid.New().String()
}

// generateRecordMetaID generates a record ID for the meta table
func generateRecordMetaID(recordID string) string {
	return RECORD_META_ID_PREFIX + recordID
}
