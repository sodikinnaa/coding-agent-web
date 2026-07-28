package config

import (
	"strings"
	"testing"
)

func TestAES256GCMEncryption(t *testing.T) {
	rawToken := "wak_secret_token_1234567890_test_key"

	// 1. Test Encrypt
	encToken := EncryptToken(rawToken)
	if !strings.HasPrefix(encToken, EncPrefix) {
		t.Fatalf("Expected encrypted token to start with %s, got: %s", EncPrefix, encToken)
	}

	if encToken == rawToken {
		t.Fatalf("Encrypted token should not equal raw token")
	}

	// 2. Test Decrypt
	decToken := DecryptToken(encToken)
	if decToken != rawToken {
		t.Fatalf("Expected decrypted token to be %s, got: %s", rawToken, decToken)
	}

	// 3. Test Pass-through unencrypted token
	unenc := DecryptToken("plain_text_token")
	if unenc != "plain_text_token" {
		t.Fatalf("Expected plain text token to remain unchanged, got: %s", unenc)
	}
}
