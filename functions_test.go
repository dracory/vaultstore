package vaultstore

import (
	"strings"
	"testing"
)

func Test_randomFromGamma(t *testing.T) {
	tests := []struct {
		name   string
		length int
		gamma  string
	}{
		{
			name:   "generate 10 chars from lowercase",
			length: 10,
			gamma:  "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:   "generate 20 chars from alphanumeric",
			length: 20,
			gamma:  "abcdefghijklmnopqrstuvwxyz0123456789",
		},
		{
			name:   "generate 32 chars from full charset",
			length: 32,
			gamma:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:   "generate 1 char",
			length: 1,
			gamma:  "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := randomFromGamma(tt.length, tt.gamma)

			if len(got) != tt.length {
				t.Errorf("randomFromGamma() length = %v, want %v", len(got), tt.length)
			}

			// Verify all characters are from gamma
			for _, char := range got {
				if !strings.ContainsRune(tt.gamma, char) {
					t.Errorf("randomFromGamma() contains invalid char %q, not in gamma %q", char, tt.gamma)
				}
			}
		})
	}
}

func Test_randomFromGamma_distribution(t *testing.T) {
	// Test that distribution is reasonably uniform (rejection sampling works)
	gamma := "ab"
	length := 1000
	counts := make(map[rune]int)

	// Generate many samples
	for i := 0; i < 100; i++ {
		s := randomFromGamma(length, gamma)
		for _, char := range s {
			counts[char]++
		}
	}

	total := 100 * length
	expected := total / len(gamma)
	tolerance := float64(expected) * 0.2 // 20% tolerance

	for char, count := range counts {
		diff := float64(count - expected)
		if diff < 0 {
			diff = -diff
		}
		if diff > tolerance {
			t.Logf("Character %q count %d differs from expected %d by %.2f%%", char, count, expected, (diff/float64(expected))*100)
		}
	}
}

func Test_generateToken(t *testing.T) {
	tests := []struct {
		name        string
		tokenLength int
	}{
		{
			name:        "generateToken of minimum valid length",
			tokenLength: len(TOKEN_PREFIX) + 12,
		},
		{
			name:        "generateToken of length 20",
			tokenLength: 20,
		},
		{
			name:        "generateToken of length 30",
			tokenLength: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateToken(tt.tokenLength)

			if err != nil {
				t.Errorf("generateToken() error = %v", err)
				return
			}
			if len(got) != tt.tokenLength {
				t.Errorf("generateToken() got = %v, want %v", len(got), tt.tokenLength)
			}

			if got[:len(TOKEN_PREFIX)] != TOKEN_PREFIX {
				t.Errorf("generateToken() got = %v, want %v", got[:len(TOKEN_PREFIX)], TOKEN_PREFIX)
			}

			if !IsToken(got) {
				t.Errorf("generateToken() got = %v, want %v", IsToken(got), true)
			}
		})
	}
}

func Test_generateToken_invalidLength(t *testing.T) {
	const minPayloadLength = 12
	minTotalLength := len(TOKEN_PREFIX) + minPayloadLength

	tests := []struct {
		name        string
		tokenLength int
	}{
		{
			name:        "generateToken of length 1 should error",
			tokenLength: 1,
		},
		{
			name:        "generateToken of length equal to prefix should error",
			tokenLength: len(TOKEN_PREFIX),
		},
		{
			name:        "generateToken of just below minimum total length should error",
			tokenLength: minTotalLength - 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateToken(tt.tokenLength)

			if err == nil {
				t.Errorf("generateToken() expected error for length %d, got token %q", tt.tokenLength, got)
			}
		})
	}
}
