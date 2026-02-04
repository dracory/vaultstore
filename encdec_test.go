package vaultstore

import "testing"

func Test_xorEncrypt(t *testing.T) {
	str := xorEncrypt("input", "key")
	if len(str) == 0 {
		t.Fatalf("xorEncrypt Failure")
	}
}

func Test_xorDecrypt(t *testing.T) {
	str := xorEncrypt("input", "key")
	out, err := xorDecrypt(str, "key")
	if err != nil {
		t.Fatalf("xorDecrypt Failure")
	}
	if out != "input" {
		t.Fatalf("xorDecrypt Failure: Expected [input] Received [%v]", out)
	}
}

func Test_xorEncryptDecryptLargeInput(t *testing.T) {
	var input = createRandomBlock(1000000)

	if len(input) != 1000000 {
		t.Fatalf("createRandomBlock Failure")
	}

	str := xorEncrypt(input, "key")

	out, err := xorDecrypt(str, "key")

	if err != nil {
		t.Fatalf("xorDecrypt Failure")
	}

	if out != input {
		t.Fatalf("xorDecrypt Failure: Expected [input] Received [%v]", out)
	}
}

func Test_decode(t *testing.T) {
	test_val := "test_value"
	test_pass := "test_password"
	encoded_str, err := encode(test_val, test_pass, nil)
	if err != nil {
		t.Fatalf("encode() failed: %v", err)
	}

	str, err := decode(encoded_str, test_pass, nil)
	if err != nil {
		t.Fatalf("decode Failure [%v]", err.Error())
	}
	if str != test_val {
		t.Fatalf("decoded String Match Failure: Expected [%v], received [%v]", test_val, str)
	}
}

func Test_encode(t *testing.T) {
	test_val := "test_value"
	test_pass := "test_password"
	encoded_str, err := encode(test_val, test_pass, nil)
	if err != nil {
		t.Fatalf("encode() failed: %v", err)
	}

	str, err := decode(encoded_str, test_pass, nil)
	if err != nil {
		t.Fatalf("encode Failure [%v]", err.Error())
	}
	if str != test_val {
		t.Fatalf("encoded String Match Failure: Expected [%v], received [%v]", test_val, str)
	}
}
