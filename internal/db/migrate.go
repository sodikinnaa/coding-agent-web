package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// RunMigrations executes all .up.sql migration files in the migrations directory
func RunMigrations(dbPath string) error {
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}
	defer database.Close()

	// Ensure schema_migrations table exists
	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	migrationDir := "migrations"
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		// Fallback to absolute path or directory relative to main
		execPath, _ := os.Getwd()
		migrationDir = filepath.Join(execPath, "migrations")
	}

	files, err := filepath.Glob(filepath.Join(migrationDir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		log.Println("[Migrate] No migration files found in migrations/, running fallback embedded schema...")
		return RunFallbackMigration(database)
	}

	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.Split(filename, "_")[0]

		var exists int
		err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if exists > 0 {
			fmt.Printf("[Migrate] Migration %s already applied. Skipping.\n", filename)
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		fmt.Printf("[Migrate] Applying migration %s...\n", filename)
		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing migration %s: %w", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", filename, err)
		}

		fmt.Printf("[Migrate] Migration %s applied successfully.\n", filename)
	}

	return nil
}

// RunRollback executes the latest .down.sql migration file
func RunRollback(dbPath string) error {
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}
	defer database.Close()

	migrationDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationDir, "*.down.sql"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no .down.sql migration files found")
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.Split(filename, "_")[0]

		var exists int
		err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err != nil || exists == 0 {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read rollback file %s: %w", file, err)
		}

		fmt.Printf("[Migrate] Rolling back migration %s...\n", filename)
		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing rollback %s: %w", filename, err)
		}

		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback for %s: %w", filename, err)
		}

		fmt.Printf("[Migrate] Rollback %s completed successfully.\n", filename)
		break
	}

	return nil
}

// RunFallbackMigration runs programmatic table creation if SQL migration files are unavailable
func RunFallbackMigration(database *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		full_name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		daily_limit INTEGER NOT NULL DEFAULT 5,
		used_today INTEGER NOT NULL DEFAULT 0,
		last_active_date TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS chat_sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS quiz_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		grade TEXT NOT NULL,
		score INTEGER NOT NULL,
		total_questions INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS custom_quizzes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		grade TEXT NOT NULL,
		topic TEXT NOT NULL,
		question TEXT NOT NULL,
		options_json TEXT NOT NULL,
		correct_index INTEGER NOT NULL,
		explanation TEXT NOT NULL,
		reference_book TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS quiz_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		grade TEXT NOT NULL,
		selected_books_json TEXT NOT NULL,
		total_questions INTEGER NOT NULL DEFAULT 5,
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS payment_transactions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		tier_name TEXT NOT NULL,
		daily_limit INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		qr_url TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		paid_at DATETIME,
		expired_at DATETIME,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS credit_packages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		daily_limit INTEGER NOT NULL,
		price INTEGER NOT NULL,
		description TEXT DEFAULT '',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := database.Exec(schema)
	return err
}
