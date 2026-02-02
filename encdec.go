package vaultstore

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"errors"
	"io"
	"math/rand/v2"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

func decode(value string, password string) (string, error) {
	// Check for v2 encryption prefix (AES-GCM)
	if strings.HasPrefix(value, ENCRYPTION_PREFIX_V2) {
		return decodeV2(value, password)
	}

	// Legacy v1 decryption (XOR-based)
	return decodeV1(value, password)
}

// decodeV1 handles legacy XOR-based decryption
func decodeV1(value string, password string) (string, error) {
	strongPassword := strongifyPassword(password)
	first, err := xorDecrypt(value, strongPassword)

	if err != nil {
		return "", errors.New("xor. " + err.Error())
	}

	if !isBase64(first) {
		return "", errors.New("decryption failed")
	}

	v4, err := base64Decode(first)

	if err != nil {
		return "", errors.New("base64.1. " + err.Error())
	}

	parts := strings.Split(string(v4), "_")

	if len(parts) < 2 {
		return "", errors.New("decryption failed")
	}

	upTo, err := strconv.Atoi(parts[0])

	if err != nil {
		return "", errors.New("atoi. " + err.Error())
	}

	after := strings.Join(parts[1:], "_")

	v1 := after[0:upTo]

	v2, err := base64Decode(v1)
	if err != nil {
		return "", errors.New("base64.2. " + err.Error())
	}

	return string(v2), nil
}

// decodeV2 handles AES-GCM decryption with Argon2id key derivation
func decodeV2(value string, password string) (string, error) {
	// Remove the v2: prefix
	encodedData := strings.TrimPrefix(value, ENCRYPTION_PREFIX_V2)

	// Decode base64
	data, err := base64Decode(encodedData)
	if err != nil {
		return "", errors.New("base64 decode: " + err.Error())
	}

	// Check minimum length (salt + nonce + tag)
	if len(data) < V2_SALT_SIZE+V2_NONCE_SIZE+V2_TAG_SIZE {
		return "", errors.New("invalid ciphertext length")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:V2_SALT_SIZE]
	nonce := data[V2_SALT_SIZE : V2_SALT_SIZE+V2_NONCE_SIZE]
	ciphertext := data[V2_SALT_SIZE+V2_NONCE_SIZE:]

	// Derive key using Argon2id
	key := deriveKeyArgon2id(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("aes cipher: " + err.Error())
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("gcm: " + err.Error())
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: " + err.Error())
	}

	return string(plaintext), nil
}

func encode(value string, password string) string {
	// Always use v2 encryption for new data
	return encodeV2(value, password)
}

// encodeV2 encrypts using AES-GCM with Argon2id key derivation
func encodeV2(value string, password string) string {
	// Generate random salt
	salt := make([]byte, V2_SALT_SIZE)
	if _, err := io.ReadFull(cryptorand.Reader, salt); err != nil {
		// Fall back to insecure random only if crypto/rand fails
		for i := range salt {
			salt[i] = byte(rand.IntN(256))
		}
	}

	// Derive key using Argon2id
	key := deriveKeyArgon2id(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		// Fall back to insecure random only if crypto/rand fails
		for i := range nonce {
			nonce[i] = byte(rand.IntN(256))
		}
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	// Combine salt + ciphertext (which includes nonce + tag)
	combined := append(salt, ciphertext...)

	// Encode and add prefix
	return ENCRYPTION_PREFIX_V2 + base64Encode(combined)
}

// deriveKeyArgon2id derives a key using Argon2id
func deriveKeyArgon2id(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt,
		ARGON2_ITERATIONS,
		ARGON2_MEMORY,
		ARGON2_PARALLELISM,
		ARGON2_KEY_LENGTH)
}

// strongifyPassword Performs multiple calculations
// on top of the password and changes it to a derivative
// long hash. This is done so that even simple and not-long
// passwords  can become longer and stronger (144 characters).
func strongifyPassword(password string) string {
	p1 := strToMD5Hash(password) + strToMD5Hash(password) + strToMD5Hash(password) + strToMD5Hash(password)

	p1 = strToSHA256Hash(p1)
	p2 := strToSHA1Hash(p1) + strToSHA1Hash(p1) + strToSHA1Hash(p1)
	p3 := strToSHA1Hash(p2) + strToMD5Hash(p2) + strToSHA1Hash(p2)
	p4 := strToSHA256Hash(p3)
	p5 := strToSHA1Hash(p4) + strToSHA256Hash(strToMD5Hash(p4)) + strToSHA256Hash(strToSHA1Hash(p4)) + strToMD5Hash(p4)
	return p5
}

// createRandomBlock returns a random string of specified length
func createRandomBlock(length int) string {
	const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = characters[rand.IntN(len(characters))]
	}
	return string(result)
}

// calculateRequiredBlockLength calculates block length (128) required to contain a length
func calculateRequiredBlockLength(v int) int {
	a := 128
	for v > a {
		a = a * 2
	}
	return a
}
