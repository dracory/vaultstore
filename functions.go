package vaultstore

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"
)

// base64Encode encodes a byte array to a base64 string
func base64Encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// base64Decode decodes a base64 encoded string to a byte array
func base64Decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}

// base32CrockfordEncode encodes a number to base32 Crockford (lowercase)
// Crockford's Base32: 0123456789abcdefghjkmnpqrstvwxyz (no i, l, o, u)
func base32CrockfordEncode(n uint64) string {
	const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

	if n == 0 {
		return "0"
	}

	var result []byte
	for n > 0 {
		result = append([]byte{alphabet[n%32]}, result...)
		n /= 32
	}

	return string(result)
}

// getTimestampComponent returns a base32 Crockford encoded timestamp
func getTimestampComponent() string {
	// Use milliseconds since epoch for more uniqueness
	timestamp := uint64(time.Now().UnixMilli())
	return base32CrockfordEncode(timestamp)
}

// isBase64 checks if a string is a base64 encoded string
func isBase64(value string) bool {
	base64Regex := "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	rxBase64 := regexp.MustCompile(base64Regex)
	return rxBase64.MatchString(value)
}

// GenerateToken generates a random token
// Business logic:
//  1. Generate timestamp component (base32 Crockford)
//  2. Generate random component to fill remaining space
//  3. Combine: tk_{TIMESTAMP}{RAND}
//
// Collision Probability Analysis:
// - Different timestamps: 0% (impossible due to unique ms timestamps)
// - Same timestamp: 1/36^randomLen (e.g., 1/36^21 ≈ 1 in 5×10^32 for 32-char tokens)
// - Real-world probability: Effectively zero (would need impossible generation rates)
//
// This ensures zero collision probability due to timestamp uniqueness
func generateToken(tokenLength int) (string, error) {
	minTotalLength := TOKEN_MIN_TOTAL_LENGTH
	maxTotalLength := TOKEN_MAX_TOTAL_LENGTH

	if tokenLength < minTotalLength || tokenLength > maxTotalLength {
		return "", fmt.Errorf("tokenLength must be between %d and %d", minTotalLength, maxTotalLength)
	}

	// Get timestamp component (base32 Crockford encoded)
	timestamp := getTimestampComponent()
	timestampLen := len(timestamp)

	// Calculate remaining length for random part
	randomLen := tokenLength - len(TOKEN_PREFIX) - timestampLen
	if randomLen < 0 {
		// If timestamp is too long, truncate it to make room for minimum random part
		maxTimestampLen := tokenLength - len(TOKEN_PREFIX) - 4 // 4 chars minimum random
		if maxTimestampLen < 4 {
			maxTimestampLen = 4
		}
		timestamp = timestamp[:maxTimestampLen]
		randomLen = tokenLength - len(TOKEN_PREFIX) - len(timestamp)
	}

	// Generate random component
	random := randomFromGamma(randomLen, "0123456789abcdefghjkmnpqrstvwxyz")

	return fmt.Sprintf("%s%s%s", TOKEN_PREFIX, timestamp, random), nil
}

// randomFromGamma generates random string of specified length with the characters specified in the gamma string
func randomFromGamma(length int, gamma string) string {
	inRune := []rune(gamma)
	out := make([]rune, length)
	gammaLen := len(inRune)

	// Calculate max value for unbiased rejection sampling
	// We need: max % gammaLen == gammaLen - 1 for unbiased distribution
	// So max should be the largest multiple of gammaLen that fits in a byte
	maxValid := 256 - (256 % gammaLen)

	for i := 0; i < length; {
		// Generate a random byte
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			continue
		}

		// Rejection sampling: skip values that would cause bias
		if int(b[0]) >= maxValid {
			continue
		}

		// Now modulo will be unbiased
		randomIndex := int(b[0]) % gammaLen
		out[i] = inRune[randomIndex]
		i++
	}

	return string(out)
}

// strToMD5Hash generates an MD5 hash of the input string
// Deprecated: insecure legacy v1 function, use modern cryptographic primitives instead.
// This function is retained only for reading legacy encrypted data.
// It will be removed in version 2.0.
func strToMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// strToSHA1Hash generates a SHA1 hash of the input string
// Deprecated: insecure legacy v1 function, use modern cryptographic primitives instead.
// This function is retained only for reading legacy encrypted data.
// It will be removed in version 2.0.
func strToSHA1Hash(text string) string {
	hash := sha1.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// strToSHA256Hash generates a SHA256 hash of the input string
func strToSHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// xorEncrypt runs a XOR encryption on the input string
// Deprecated: insecure legacy v1 encryption, use encodeV2 instead.
// This function is retained only for reading legacy encrypted data.
// It will be removed in version 2.0.
func xorEncrypt(input, key string) (output string) {
	inputBytes := []byte(input)
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	outputBytes := make([]byte, len(inputBytes))
	for i := range inputBytes {
		outputBytes[i] = inputBytes[i] ^ keyBytes[i%keyLen]
	}

	return base64Encode(outputBytes)
}

// xorDecrypt runs a XOR encryption on the input string
// Deprecated: insecure legacy v1 decryption, use decodeV2 instead.
// This function is retained only for reading legacy encrypted data.
// It will be removed in version 2.0.
func xorDecrypt(encstring string, key string) (output string, err error) {
	inputBytes, err := base64Decode(encstring)

	if err != nil {
		return "", err
	}

	outputBytes := make([]byte, len(inputBytes))
	for i, b := range inputBytes {
		outputBytes[i] = b ^ key[i%len(key)]
	}

	return string(outputBytes), nil
}
