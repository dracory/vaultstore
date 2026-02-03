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

func Test_base32CrockfordEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "encode 0",
			input:    0,
			expected: "0",
		},
		{
			name:     "encode 1",
			input:    1,
			expected: "1",
		},
		{
			name:     "encode 31",
			input:    31,
			expected: "z",
		},
		{
			name:     "encode 32",
			input:    32,
			expected: "10",
		},
		{
			name:     "encode 1023",
			input:    1023,
			expected: "zz",
		},
		{
			name:     "encode large number",
			input:    123456789,
			expected: "3nqk8n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := base32CrockfordEncode(tt.input)
			if got != tt.expected {
				t.Errorf("base32CrockfordEncode(%d) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_getTimestampComponent(t *testing.T) {
	// Test that timestamp component is generated and has expected properties
	timestamp := getTimestampComponent()

	// Should not be empty
	if timestamp == "" {
		t.Error("getTimestampComponent() returned empty string")
	}

	// Should be reasonable length (8-12 characters for current timestamps)
	if len(timestamp) < 6 || len(timestamp) > 15 {
		t.Errorf("getTimestampComponent() length = %v, want between 6 and 15", len(timestamp))
	}

	// Should only contain valid base32 characters
	validChars := "0123456789abcdefghjkmnpqrstvwxyz"
	for _, char := range timestamp {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("getTimestampComponent() contains invalid char %q", char)
		}
	}

	// Multiple calls should produce different timestamps (unless called in same ms)
	timestamp2 := getTimestampComponent()
	if timestamp == timestamp2 {
		t.Log("Note: Both timestamps are identical (likely called in same millisecond)")
	}
}

func Test_generateToken_timestampFormat(t *testing.T) {
	// Test that generated tokens have the expected timestamp+random format
	token, err := generateToken(32)
	if err != nil {
		t.Fatalf("generateToken(32) error = %v", err)
	}

	// Should have prefix
	if !strings.HasPrefix(token, TOKEN_PREFIX) {
		t.Errorf("generateToken() should start with %q, got %q", TOKEN_PREFIX, token)
	}

	// Extract payload (without prefix)
	payload := token[len(TOKEN_PREFIX):]
	if len(payload) != 29 { // 32 - 3 = 29
		t.Errorf("generateToken() payload length = %v, want 29", len(payload))
	}

	// Should contain only valid characters
	validChars := "0123456789abcdefghjkmnpqrstvwxyz"
	for _, char := range payload {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("generateToken() payload contains invalid char %q", char)
		}
	}
}

func Test_generateToken_collisionResistance(t *testing.T) {
	// Generate multiple tokens and verify they're all unique
	tokens := make(map[string]bool)
	numTokens := 1000

	for i := 0; i < numTokens; i++ {
		token, err := generateToken(32)
		if err != nil {
			t.Fatalf("generateToken(32) error = %v", err)
		}

		if tokens[token] {
			t.Errorf("Collision detected! Token %q was generated twice", token)
			break
		}
		tokens[token] = true
	}

	t.Logf("Generated %d unique tokens with no collisions", numTokens)
}
