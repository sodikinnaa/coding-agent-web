package auth

import (
	"testing"
)

func TestRegisterUserValidation(t *testing.T) {
	_, err := RegisterUser("", "password123", "Test User")
	if err == nil {
		t.Errorf("Expected error for empty username, got nil")
	}

	_, err = RegisterUser("testuser", "", "Test User")
	if err == nil {
		t.Errorf("Expected error for empty password, got nil")
	}
}

func TestLogoutUser(t *testing.T) {
	// Test clearing non-existent or existing token
	LogoutUser("token_test_123")
	sessionMutex.RLock()
	_, exists := sessions["token_test_123"]
	sessionMutex.RUnlock()

	if exists {
		t.Errorf("Expected session token_test_123 to be deleted")
	}
}
