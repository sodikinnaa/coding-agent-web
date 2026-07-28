package main

import (
	"flag"
	"fmt"
	"log"

	"coding_agent_web/internal/db"
)

func main() {
	dbPath := flag.String("db", "data.db", "Path to SQLite database file")
	action := flag.String("action", "up", "Migration action: 'up' (apply pending migrations) or 'down' (rollback last migration)")
	flag.Parse()

	fmt.Printf("=== Database Migration Tool ===\n")
	fmt.Printf("Database Path : %s\n", *dbPath)
	fmt.Printf("Action        : %s\n", *action)

	switch *action {
	case "up":
		if err := db.RunMigrations(*dbPath); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migration completed successfully.")
	case "down":
		if err := db.RunRollback(*dbPath); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✓ Rollback completed successfully.")
	default:
		log.Fatalf("Unknown action '%s'. Supported actions: 'up', 'down'", *action)
	}
}
