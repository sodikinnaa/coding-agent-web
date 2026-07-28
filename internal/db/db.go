package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open sqlite database: %v", err)
	}

	if err := RunMigrations(dbPath); err != nil {
		log.Printf("Warning: RunMigrations encounter: %v. Running fallback schema...", err)
		if err := RunFallbackMigration(DB); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
	}

	fmt.Println("SQLite Database initialized successfully.")
}
