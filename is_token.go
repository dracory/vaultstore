package vaultstore

import "strings"

func IsToken(s string) bool {
	return strings.HasPrefix(s, TOKEN_PREFIX)
}

// IsTokenValidLength checks if a token has valid format and reasonable length
// Returns false if token format is invalid or length is outside reasonable bounds
func IsTokenValidLength(s string) bool {
	if len(s) < TOKEN_MIN_TOTAL_LENGTH || len(s) > TOKEN_MAX_TOTAL_LENGTH {
		return false
	}

	return IsToken(s)
}
