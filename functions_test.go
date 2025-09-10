package vaultstore

import "testing"

func Test_generateToken(t *testing.T) {
	tests := []struct {
		name        string
		tokenLength int
	}{
		{
			name:        "generateToken of length 10",
			tokenLength: 10,
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
