package vaultstore

import (
	"testing"
)

var test_val = createRandomBlock(1000000)

// var test_val = randomFromGamma(100000, "abcdefghijklmnopqrstuvwxyz0123456789")

func BenchmarkEnc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := encode(test_val, "test_password", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
