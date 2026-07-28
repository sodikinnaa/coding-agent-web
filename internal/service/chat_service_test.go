package service

import (
	"testing"
)

func TestMaskIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sodikinnaa@gmail.com", "so***aa@gmail.com"},
		{"john.doe@example.org", "jo***oe@example.org"},
		{"ab@domain.com", "a***@domain.com"},
		{"sodikinnaa", "sod***aa"},
		{"budi", "b***i"},
		{"ab", "a***"},
		{"", "user***"},
	}

	for _, tt := range tests {
		result := maskIdentifier(tt.input)
		if result != tt.expected {
			t.Errorf("maskIdentifier(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetFallbackQuiz(t *testing.T) {
	questions := getFallbackQuiz("Kelas 5 SD")
	if len(questions) == 0 {
		t.Fatalf("Expected fallback quiz questions, got empty slice")
	}

	for i, q := range questions {
		if q.Question == "" {
			t.Errorf("Question at index %d has empty question text", i)
		}
		if len(q.Options) != 4 {
			t.Errorf("Question at index %d expected 4 options, got %d", i, len(q.Options))
		}
		if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
			t.Errorf("Question at index %d invalid CorrectIndex: %d", i, q.CorrectIndex)
		}
	}
}

func TestQuizSessionAndVerification(t *testing.T) {
	sessionID, clientQuestions, err := CreateQuizSession(100, 0, "Kelas 5 SD")
	if err != nil {
		t.Fatalf("CreateQuizSession failed: %v", err)
	}

	if sessionID == "" {
		t.Errorf("Expected non-empty sessionID")
	}

	if len(clientQuestions) == 0 {
		t.Fatalf("Expected client questions, got empty slice")
	}

	// Verify first question
	isCorrect, correctIndex, explanation, err := VerifyQuizAnswer(sessionID, 0, 0)
	if err != nil {
		t.Fatalf("VerifyQuizAnswer failed: %v", err)
	}

	if correctIndex < 0 || correctIndex > 3 {
		t.Errorf("Invalid correctIndex returned: %d", correctIndex)
	}

	if explanation == "" {
		t.Errorf("Expected non-empty explanation")
	}

	if isCorrect != (0 == correctIndex) {
		t.Errorf("isCorrect mismatch: got %v, want %v", isCorrect, (0 == correctIndex))
	}
}
