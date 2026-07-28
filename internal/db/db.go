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

	seedDefaultCategories(DB)

	fmt.Println("SQLite Database initialized successfully.")
}

func seedDefaultCategories(database *sql.DB) {
	var count int
	_ = database.QueryRow("SELECT COUNT(*) FROM quiz_categories").Scan(&count)
	if count <= 1 {
		_, _ = database.Exec(`
			INSERT OR IGNORE INTO quiz_categories (id, name, grade, selected_books_json, total_questions, description) VALUES
			(1, 'Kuis AI & Pemrograman SD Kelas 5', 'Kelas 5 SD', '["BUKU AI SD-MI Kelas 5 Semester 1.pdf","BUKU AI SD-MI Kelas 5 Semester 2.pdf"]', 5, 'Kuis evaluasi pemahaman AI, etika digital, dan pemrograman visual Scratch untuk siswa SD Kelas 5.'),
			(2, 'Kuis AI & Pemrograman SD Kelas 6', 'Kelas 6 SD', '["BUKU AI SD-MI Kelas 6 Semester 1.pdf","BUKU AI SD-MI Kelas 6 Semester 2.pdf"]', 5, 'Kuis algoritma dasar, berpikir komputasional, dan aplikasi AI untuk siswa SD Kelas 6.'),
			(3, 'Kuis Coding & AI SMA Kelas 10', 'Kelas 10 SMA', '["BUKU AI SMA Kelas 10 Semester 1.pdf","BUKU AI SMA Kelas 10 Semester 2.pdf"]', 5, 'Kuis logika pemrograman, algoritma dasar Python, dan pengenalan Artificial Intelligence untuk SMA Kelas 10.'),
			(4, 'Kuis Coding & AI SMA Kelas 11', 'Kelas 11 SMA', '["BUKU AI SMA Kelas 11 Semester 1.pdf"]', 5, 'Kuis struktur data, pemrosesan data AI, dan jaringan saraf tiruan dasar untuk SMA Kelas 11.'),
			(5, 'Kuis Dasar-Dasar Pemrograman', 'Coding Dasar', '["( 8 ) CODING DASAR BRO..pdf"]', 5, 'Kuis konsep dasar coding, variabel, kondisional, dan perulangan.'),
			(6, 'Kuis Kombinatorik & Logika Koding', 'Kombinatorik', '["( 8 ) CODING KOMBINATORIK..pdf"]', 5, 'Kuis logika matematika, permutasi, kombinasi, dan problem solving algoritma.'),
			(7, 'Kuis Matematika Koding & Algoritma', 'Coding Matematika', '["( 8 ) CODING MTK..pdf"]', 5, 'Kuis penerapan matematika komputasional dan operasi logika pemrograman.'),
			(8, 'Kuis Coding SD Fase C (Silabus & CP)', 'SD Fase C', '["( 8 ) CODING SD FASE C.pdf"]', 5, 'Kuis capaian pembelajaran koding dan dasar-dasar computational thinking SD Fase C.');
		`)
	}
}
