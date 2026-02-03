package vaultstore

import "testing"

func TestIsToken(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is token",
			args: args{
				s: "tk_12345",
			},
			want: true,
		},
		{
			name: "is not token",
			args: args{
				s: "tkn_123456",
			},
			want: false,
		},
		{
			name: "is not token",
			args: args{
				s: "12345",
			},
			want: false,
		},
		{
			name: "is not token",
			args: args{
				s: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsToken(tt.args.s); got != tt.want {
				t.Errorf("IsToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTokenValidLength(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid token with minimum length",
			args: args{
				s: "tk_123456789012",
			},
			want: true,
		},
		{
			name: "valid token with reasonable length",
			args: args{
				s: "tk_12345678901234567890123456789012", // 35 chars total
			},
			want: true,
		},
		{
			name: "token too short",
			args: args{
				s: "tk_123",
			},
			want: false,
		},
		{
			name: "token too long",
			args: args{
				s: "tk_" + string(make([]byte, 50)), // 53 chars total, exceeds 35 max
			},
			want: false,
		},
		{
			name: "wrong format but valid length",
			args: args{
				s: "abc_12345678901234567890",
			},
			want: false,
		},
		{
			name: "empty string",
			args: args{
				s: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTokenValidLength(tt.args.s); got != tt.want {
				t.Errorf("IsTokenValidLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
