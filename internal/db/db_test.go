package db

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	testDBPath := "./test_data.db"
	defer os.Remove(testDBPath)

	InitDB(testDBPath)

	if DB == nil {
		t.Fatalf("Expected DB connection to be initialized, got nil")
	}

	// Verify tables exist
	var tableName string
	err := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err != nil || tableName != "users" {
		t.Errorf("Expected 'users' table to exist in SQLite database")
	}
}
