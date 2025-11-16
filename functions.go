package vaultstore

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
)

// base64Encode encodes a byte array to a base64 string
func base64Encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// base64Decode decodes a base64 encoded string to a byte array
func base64Decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}

// fileExists checks if a file exists
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return !os.IsNotExist(err)
}

// isBase64 checks if a string is a base64 encoded string
func isBase64(value string) bool {
	base64Regex := "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	rxBase64 := regexp.MustCompile(base64Regex)
	return rxBase64.MatchString(value)
}

// GenerateToken generates a random token
// Business logic:
//  1. Generate random lowercase string
//  2. Prefix with "tk_"
func generateToken(tokenLength int) (string, error) {
	const minPayloadLength = 12
	minTotalLength := len(TOKEN_PREFIX) + minPayloadLength
	if tokenLength < minTotalLength {
		return "", fmt.Errorf("tokenLength must be at least %d (prefix %d + payload %d)", minTotalLength, len(TOKEN_PREFIX), minPayloadLength)
	}

	token := randomFromGamma(tokenLength-len(TOKEN_PREFIX), "abcdefghijklmnopqrstuvwxyz0123456789")

	return fmt.Sprintf("%s%s", TOKEN_PREFIX, token), nil
}

// randomFromGamma generates random string of specified length with the characters specified in the gamma string
func randomFromGamma(length int, gamma string) string {
	inRune := []rune(gamma)
	out := ""

	for i := 0; i < length+20; i++ {
		// Generate a random byte
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			continue
		}

		// Map the byte to an index in the gamma string
		randomIndex := int(b[0]) % len(inRune)
		pick := inRune[randomIndex]
		out += string(pick)
	}

	return out[:length]
}

// strToMD5Hash generates an MD5 hash of the input string
func strToMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// strToSHA1Hash generates a SHA1 hash of the input string
func strToSHA1Hash(text string) string {
	hash := sha1.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// strToSHA256Hash generates a SHA256 hash of the input string
func strToSHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// xorEncrypt  runs a XOR encryption on the input string
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

// xorDecrypt  runs a XOR encryption on the input string
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
