package service

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"coding_agent_web/internal/config"
	"coding_agent_web/internal/db"
	"coding_agent_web/internal/knowledge"
	"coding_agent_web/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type QuizQuestion struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectIndex  int      `json:"correct_index"`
	Explanation   string   `json:"explanation"`
	ReferenceBook string   `json:"reference_book"`
}

type ClientQuizQuestion struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	ReferenceBook string   `json:"reference_book"`
}

type QuizSessionData struct {
	ID        string
	UserID    int64
	Grade     string
	Questions []QuizQuestion
	CreatedAt time.Time
}

var (
	quizSessions      = make(map[string]*QuizSessionData)
	quizSessionsMutex sync.RWMutex
)

func CreateQuizSession(userID int64, categoryID int64, grade string) (string, []ClientQuizQuestion, error) {
	questions, err := GenerateQuizByCategory(categoryID, grade)
	if err != nil || len(questions) == 0 {
		questions = getFallbackQuiz(grade)
	}

	sessionID := fmt.Sprintf("qz_%d_%d", time.Now().UnixNano(), userID)

	quizSessionsMutex.Lock()
	quizSessions[sessionID] = &QuizSessionData{
		ID:        sessionID,
		UserID:    userID,
		Grade:     grade,
		Questions: questions,
		CreatedAt: time.Now(),
	}
	quizSessionsMutex.Unlock()

	clientQuestions := make([]ClientQuizQuestion, len(questions))
	for i, q := range questions {
		clientQuestions[i] = ClientQuizQuestion{
			ID:            q.ID,
			Question:      q.Question,
			Options:       q.Options,
			ReferenceBook: q.ReferenceBook,
		}
	}

	return sessionID, clientQuestions, nil
}

func VerifyQuizAnswer(sessionID string, qIndex int, selectedIndex int) (bool, int, string, error) {
	quizSessionsMutex.RLock()
	session, exists := quizSessions[sessionID]
	quizSessionsMutex.RUnlock()

	if !exists {
		return false, 0, "", fmt.Errorf("Sesi kuis tidak ditemukan atau telah kedaluwarsa")
	}

	if qIndex < 0 || qIndex >= len(session.Questions) {
		return false, 0, "", fmt.Errorf("Indeks soal tidak valid")
	}

	q := session.Questions[qIndex]
	isCorrect := (selectedIndex == q.CorrectIndex)

	return isCorrect, q.CorrectIndex, q.Explanation, nil
}

// In-memory cache for Base64 encoded document content to avoid re-reading and re-encoding on every request
var (
	docBase64Cache = make(map[string]string)
	cacheMutex     sync.RWMutex
)

func getDocBase64(path string) (string, error) {
	cacheMutex.RLock()
	data, exists := docBase64Cache[path]
	cacheMutex.RUnlock()
	if exists {
		return data, nil
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(fileBytes)

	cacheMutex.Lock()
	docBase64Cache[path] = encoded
	cacheMutex.Unlock()

	return encoded, nil
}

// getRelevantDocParts picks relevant PDF/DOCX documents from /root/coding_dataset based on prompt keywords, or all small ones
func getRelevantDocParts(prompt string) []model.ContentPart {
	docList := knowledge.GetDocList()
	if len(docList) == 0 {
		return nil
	}

	promptLower := strings.ToLower(prompt)

	var selectedFiles []string
	for _, doc := range docList {
		filenameLower := strings.ToLower(doc.Filename)
		// Check for keyword matches in filename
		if strings.Contains(promptLower, "sd") && strings.Contains(filenameLower, "sd") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "smp") && strings.Contains(filenameLower, "smp") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "sma") && strings.Contains(filenameLower, "sma") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "kelas 5") && strings.Contains(filenameLower, "kelas 5") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "kelas 6") && strings.Contains(filenameLower, "kelas 6") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "kelas 10") && strings.Contains(filenameLower, "kelas 10") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "kelas 11") && strings.Contains(filenameLower, "kelas 11") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "kombinatorik") && strings.Contains(filenameLower, "kombinatorik") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "mtk") && strings.Contains(filenameLower, "mtk") {
			selectedFiles = append(selectedFiles, doc.Filename)
		} else if strings.Contains(promptLower, "dasar") && strings.Contains(filenameLower, "dasar") {
			selectedFiles = append(selectedFiles, doc.Filename)
		}
	}

	// Default fallback: if no file matched keywords, select small core outline documents (< 2MB) to prevent payload bloat
	if len(selectedFiles) == 0 {
		for _, doc := range docList {
			filenameLower := strings.ToLower(doc.Filename)
			if strings.Contains(filenameLower, "outline") || strings.Contains(filenameLower, "cp koding") {
				selectedFiles = append(selectedFiles, doc.Filename)
				if len(selectedFiles) >= 2 {
					break
				}
			}
		}
	}
	if len(selectedFiles) == 0 && len(docList) > 0 {
		// Pick smallest document
		selectedFiles = append(selectedFiles, "BUKU AI Outline Buku Coding dan AI SD_MI.pdf")
	}

	var parts []model.ContentPart
	for _, fname := range selectedFiles {
		fullPath, mimeType, err := knowledge.GetOriginalFilePath(fname)
		if err != nil {
			continue
		}

		// Filter out non-supported MIME types (like docx) for OpenAI Vision Base64 payload
		if mimeType != "application/pdf" && !strings.HasPrefix(mimeType, "image/") {
			continue
		}

		b64Data, err := getDocBase64(fullPath)
		if err != nil {
			log.Printf("Error reading/encoding document %s: %v", fullPath, err)
			continue
		}

		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64Data)
		parts = append(parts, model.ContentPart{
			Type: "image_url",
			ImageURL: &model.ImageURL{
				URL: dataURL,
			},
		})
	}

	return parts
}

func SaveQuizScore(userID int64, grade string, score, totalQuestions int) error {
	_, err := db.DB.Exec("INSERT INTO quiz_scores (user_id, grade, score, total_questions) VALUES (?, ?, ?, ?)", userID, grade, score, totalQuestions)
	return err
}

func maskIdentifier(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "user***"
	}

	if strings.Contains(s, "@") {
		parts := strings.SplitN(s, "@", 2)
		local := parts[0]
		domain := parts[1]

		if len(local) <= 2 {
			local = string(local[0]) + "***"
		} else if len(local) <= 4 {
			local = string(local[0]) + "***" + string(local[len(local)-1])
		} else {
			local = local[:2] + "***" + local[len(local)-2:]
		}
		return local + "@" + domain
	}

	if len(s) <= 2 {
		return string(s[0]) + "***"
	} else if len(s) <= 5 {
		return string(s[0]) + "***" + string(s[len(s)-1])
	} else {
		return s[:3] + "***" + s[len(s)-2:]
	}
}

func GetLeaderboard() ([]model.LeaderboardEntry, error) {
	query := `
	SELECT u.id, u.username, u.full_name, SUM(qs.score) as total_score, COUNT(qs.id) as quiz_count, MAX(qs.grade) as high_grade
	FROM users u
	JOIN quiz_scores qs ON u.id = qs.user_id
	GROUP BY u.id
	ORDER BY total_score DESC, quiz_count DESC
	LIMIT 20
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []model.LeaderboardEntry
	for rows.Next() {
		var entry model.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.FullName, &entry.TotalScore, &entry.QuizCount, &entry.HighGrade); err == nil {
			entry.Username = maskIdentifier(entry.Username)
			if entry.FullName != "" {
				entry.FullName = maskIdentifier(entry.FullName)
			}
			leaderboard = append(leaderboard, entry)
		}
	}
	return leaderboard, nil
}

func GenerateQuizByCategory(categoryID int64, grade string) ([]QuizQuestion, error) {
	currCfg := config.GetConfig()
	totalQuestions := 5
	categoryName := "Kuis Koding & AI"

	if categoryID > 0 {
		cat, err := GetQuizCategoryByID(categoryID)
		if err == nil && cat != nil {
			if cat.TotalQuestions > 0 {
				totalQuestions = cat.TotalQuestions
			}
			categoryName = cat.Name
			if cat.Grade != "" {
				grade = cat.Grade
			}
		}
	}

	promptText := fmt.Sprintf(`Buatkan %d soal kuis pilihan ganda interaktif untuk materi "%s" (Jenjang %s) berdasarkan kurikulum Koding & AI.

PENTING: Balas HANYA dengan JSON valid tanpa markdown formatting dengan format array of objects berikut:
[
  {
    "id": 1,
    "question": "Pertanyaan soal...",
    "options": ["Pilihan A", "Pilihan B", "Pilihan C", "Pilihan D"],
    "correct_index": 1,
    "explanation": "Penjelasan mengapa jawaban tersebut benar...",
    "reference_book": "Nama Buku - Halaman X"
  }
]`, totalQuestions, categoryName, grade)

	docParts := getRelevantDocParts(grade + " " + categoryName)

	userContentParts := []model.ContentPart{
		{Type: "text", Text: promptText},
	}
	userContentParts = append(userContentParts, docParts...)

	messages := []model.ChatMessage{
		{Role: "system", Content: "Kamu adalah generator kuis edukasi koding & AI. Keluarkan output HANYA berupa JSON valid sesuai instruksi."},
		{Role: "user", Content: userContentParts},
	}

	openAiReq := model.OpenAIRequest{
		Model:    currCfg.Model,
		Messages: messages,
	}

	bodyBytes, _ := json.Marshal(openAiReq)
	url := strings.TrimSuffix(currCfg.BaseURL, "/") + "/chat/completions"

	httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	if currCfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+currCfg.APIKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return getFallbackQuiz(grade), nil
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var openAiResp model.OpenAIResponse
	if err := json.Unmarshal(respBytes, &openAiResp); err != nil || len(openAiResp.Choices) == 0 {
		return getFallbackQuiz(grade), nil
	}

	var rawReply string
	switch v := openAiResp.Choices[0].Message.Content.(type) {
	case string:
		rawReply = v
	default:
		b, _ := json.Marshal(v)
		rawReply = string(b)
	}

	rawReply = strings.TrimPrefix(rawReply, "```json")
	rawReply = strings.TrimPrefix(rawReply, "```")
	rawReply = strings.TrimSuffix(rawReply, "```")
	rawReply = strings.TrimSpace(rawReply)

	var questions []QuizQuestion
	if err := json.Unmarshal([]byte(rawReply), &questions); err != nil {
		return getFallbackQuiz(grade), nil
	}

	return questions, nil
}

func getFallbackQuiz(grade string) []QuizQuestion {
	return []QuizQuestion{
		{
			ID:            1,
			Question:      "Apa yang dimaksud dengan coding (pemrograman)?",
			Options:       []string{"Menggambar gambar digital di komputer", "Memberikan urutan perintah/instruksi kepada komputer agar menjalankan tugas tertentu", "Bermain game online sepanjang hari", "Memperbaiki komputer yang rusak"},
			CorrectIndex: 1,
			Explanation:   "Coding adalah proses memberikan instruksi atau perintah yang dapat dimengerti oleh komputer.",
			ReferenceBook: "BUKU AI SD-MI Kelas 5 Semester 1.pdf - Halaman 10",
		},
		{
			ID:            2,
			Question:      "Urutan langkah-langkah logis dan sistematis untuk menyelesaikan suatu masalah disebut...",
			Options:       []string{"Perangkat keras", "Internet", "Algoritma", "Poster"},
			CorrectIndex: 2,
			Explanation:   "Algoritma adalah langkah-langkah terstruktur untuk menyelesaikan masalah.",
			ReferenceBook: "BUKU AI SD-MI Kelas 5 Semester 1.pdf - Halaman 11",
		},
		{
			ID:            3,
			Question:      "Dalam pemrograman visual (seperti Scratch), kita membuat program dengan cara...",
			Options:       []string{"Mengetik kode angka yang rumit", "Menyusun blok-blok perintah berwarna-warni", "Memfoto layar komputer", "Mengisi formulir kertas"},
			CorrectIndex: 1,
			Explanation:   "Pemrograman visual menggunakan blok-blok kode interaktif yang disusun.",
			ReferenceBook: "BUKU AI SD-MI Kelas 5 Semester 1.pdf - Halaman 29",
		},
		{
			ID:            4,
			Question:      "JIKA hari hujan, MAKA memakai jas hujan. Konsep pemrograman yang digunakan adalah...",
			Options:       []string{"Pengulangan (Loop)", "Percabangan (If-Else)", "Bug", "Hardware"},
			CorrectIndex: 1,
			Explanation:   "Pernyataan JIKA-MAKA adalah konsep percabangan (Condition/If-Else).",
			ReferenceBook: "BUKU AI SD-MI Kelas 5 Semester 2.pdf - Halaman 10",
		},
		{
			ID:            5,
			Question:      "Manakah contoh penerapan Kecerdasan Artifisial (AI) dalam kehidupan sehari-hari?",
			Options:       []string{"Buku tulis bergaris", "Rekomendasi video YouTube / Asisten Suara", "Papan tulis kapur", "Pensil 2B"},
			CorrectIndex: 1,
			Explanation:   "Asisten suara dan rekomendasi video menggunakan algoritma AI untuk mengenali pola.",
			ReferenceBook: "BUKU AI SD-MI Kelas 5 Semester 1.pdf - Halaman 14",
		},
	}
}

// User Chat History & Sessions Management

func CreateChatSession(userID int64, firstPrompt string) (*model.ChatSession, error) {
	sessionID := generateUUID()
	title := firstPrompt
	if len(title) > 30 {
		title = title[:30] + "..."
	}
	if title == "" {
		title = "Obrolan Baru"
	}

	_, err := db.DB.Exec("INSERT INTO chat_sessions (id, user_id, title) VALUES (?, ?, ?)", sessionID, userID, title)
	if err != nil {
		return nil, err
	}

	return &model.ChatSession{
		ID:        sessionID,
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func GetUserSessions(userID int64) ([]model.ChatSession, error) {
	rows, err := db.DB.Query("SELECT id, user_id, title, created_at, updated_at FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ChatSession
	for rows.Next() {
		var s model.ChatSession
		rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
		list = append(list, s)
	}
	return list, nil
}

func GetSessionMessages(sessionID string) ([]model.ChatMessageRecord, error) {
	rows, err := db.DB.Query("SELECT id, session_id, role, content, created_at FROM chat_messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ChatMessageRecord
	for rows.Next() {
		var m model.ChatMessageRecord
		rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt)
		list = append(list, m)
	}
	return list, nil
}

func SaveChatMessage(sessionID, role, content string) error {
	_, err := db.DB.Exec("INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)", sessionID, role, content)
	if err != nil {
		return err
	}
	db.DB.Exec("UPDATE chat_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
	return nil
}

func ProcessChatWithSession(userID int64, sessionID, prompt string) (string, string, error) {
	var targetSessionID = sessionID

	if targetSessionID == "" {
		newSess, err := CreateChatSession(userID, prompt)
		if err != nil {
			return "", "", err
		}
		targetSessionID = newSess.ID
	}

	// Save User Message
	SaveChatMessage(targetSessionID, "user", prompt)

	// Fetch recent history for context (up to 10 messages)
	historyRecords, _ := GetSessionMessages(targetSessionID)

	currCfg := config.GetConfig()

	messages := []model.ChatMessage{
		{Role: "system", Content: currCfg.SystemPrompt},
	}

	// Append past text history (excluding the current last message which we will attach doc payload to)
	for i, rec := range historyRecords {
		if i == len(historyRecords)-1 && rec.Role == "user" {
			// This is the current message, handled below
			continue
		}
		messages = append(messages, model.ChatMessage{
			Role:    rec.Role,
			Content: rec.Content,
		})
	}

	// Build current user message with Direct Base64 PDF / DOCX payload
	docParts := getRelevantDocParts(prompt)
	userContentParts := []model.ContentPart{
		{Type: "text", Text: prompt},
	}
	userContentParts = append(userContentParts, docParts...)

	messages = append(messages, model.ChatMessage{
		Role:    "user",
		Content: userContentParts,
	})

	openAiReq := model.OpenAIRequest{
		Model:    currCfg.Model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(openAiReq)
	if err != nil {
		return targetSessionID, "", fmt.Errorf("Failed to encode request: %v", err)
	}

	url := strings.TrimSuffix(currCfg.BaseURL, "/") + "/chat/completions"

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return targetSessionID, "", fmt.Errorf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if currCfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+currCfg.APIKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return targetSessionID, "", fmt.Errorf("Failed to reach AI API: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return targetSessionID, "", fmt.Errorf("Failed to read response body: %v", err)
	}

	var openAiResp model.OpenAIResponse
	if err := json.Unmarshal(respBytes, &openAiResp); err != nil {
		return targetSessionID, "", fmt.Errorf("Failed to parse API response: %v", err)
	}

	if openAiResp.Error != nil {
		return targetSessionID, "", fmt.Errorf("%s", openAiResp.Error.Message)
	}

	if len(openAiResp.Choices) > 0 {
		var reply string
		switch v := openAiResp.Choices[0].Message.Content.(type) {
		case string:
			reply = v
		default:
			b, _ := json.Marshal(v)
			reply = string(b)
		}

		SaveChatMessage(targetSessionID, "assistant", reply)
		return targetSessionID, reply, nil
	}

	return targetSessionID, "", fmt.Errorf("Empty choices from AI Model")
}

func DeleteChatSession(userID int64, sessionID string) error {
	_, err := db.DB.Exec("DELETE FROM chat_sessions WHERE id = ? AND user_id = ?", sessionID, userID)
	return err
}

// Admin Service Helpers

func GetAllUsersAdmin() ([]model.AdminUserItem, error) {
	query := `
	SELECT u.id, u.username, u.full_name, u.role, u.daily_limit, u.used_today, u.created_at,
		COALESCE(COUNT(qs.id), 0) as quiz_count,
		COALESCE(SUM(qs.score), 0) as total_score
	FROM users u
	LEFT JOIN quiz_scores qs ON u.id = qs.user_id
	GROUP BY u.id
	ORDER BY u.id DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AdminUserItem
	for rows.Next() {
		var item model.AdminUserItem
		if err := rows.Scan(&item.ID, &item.Username, &item.FullName, &item.Role, &item.DailyLimit, &item.UsedToday, &item.CreatedAt, &item.QuizCount, &item.TotalScore); err == nil {
			list = append(list, item)
		}
	}
	return list, nil
}

func SetUserDailyLimitAdmin(userID int64, dailyLimit int) error {
	_, err := db.DB.Exec("UPDATE users SET daily_limit = ? WHERE id = ?", dailyLimit, userID)
	return err
}

func DeleteUserAdmin(userID int64) error {
	_, err := db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	return err
}

func ResetUserPasswordAdmin(userID int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), userID)
	return err
}

func GetCustomQuizzes(grade, topic string) ([]model.CustomQuizItem, error) {
	var rows *sql.Rows
	var err error
	if grade != "" && topic != "" {
		rows, err = db.DB.Query("SELECT id, grade, topic, question, options_json, correct_index, explanation, reference_book, created_at FROM custom_quizzes WHERE grade = ? AND topic = ? ORDER BY id DESC", grade, topic)
	} else if grade != "" {
		rows, err = db.DB.Query("SELECT id, grade, topic, question, options_json, correct_index, explanation, reference_book, created_at FROM custom_quizzes WHERE grade = ? ORDER BY id DESC", grade)
	} else if topic != "" {
		rows, err = db.DB.Query("SELECT id, grade, topic, question, options_json, correct_index, explanation, reference_book, created_at FROM custom_quizzes WHERE topic = ? ORDER BY id DESC", topic)
	} else {
		rows, err = db.DB.Query("SELECT id, grade, topic, question, options_json, correct_index, explanation, reference_book, created_at FROM custom_quizzes ORDER BY id DESC")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CustomQuizItem
	for rows.Next() {
		var q model.CustomQuizItem
		var optJSON string
		if err := rows.Scan(&q.ID, &q.Grade, &q.Topic, &q.Question, &optJSON, &q.CorrectIndex, &q.Explanation, &q.ReferenceBook, &q.CreatedAt); err == nil {
			_ = json.Unmarshal([]byte(optJSON), &q.Options)
			list = append(list, q)
		}
	}
	return list, nil
}

func CreateCustomQuiz(q model.CustomQuizItem) error {
	optJSON, _ := json.Marshal(q.Options)
	_, err := db.DB.Exec("INSERT INTO custom_quizzes (grade, topic, question, options_json, correct_index, explanation, reference_book) VALUES (?, ?, ?, ?, ?, ?, ?)",
		q.Grade, q.Topic, q.Question, string(optJSON), q.CorrectIndex, q.Explanation, q.ReferenceBook)
	return err
}

func DeleteCustomQuiz(id int64) error {
	_, err := db.DB.Exec("DELETE FROM custom_quizzes WHERE id = ?", id)
	return err
}

func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}


// Quiz Categories Service

func GetQuizCategories() ([]model.QuizCategoryItem, error) {
	rows, err := db.DB.Query("SELECT id, name, grade, selected_books_json, total_questions, description, created_at FROM quiz_categories ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.QuizCategoryItem
	for rows.Next() {
		var item model.QuizCategoryItem
		var booksJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Grade, &booksJSON, &item.TotalQuestions, &item.Description, &item.CreatedAt); err == nil {
			_ = json.Unmarshal([]byte(booksJSON), &item.SelectedBooks)
			list = append(list, item)
		}
	}
	return list, nil
}

func CreateQuizCategory(item model.QuizCategoryItem) error {
	booksJSON, _ := json.Marshal(item.SelectedBooks)
	if item.TotalQuestions <= 0 {
		item.TotalQuestions = 5
	}
	_, err := db.DB.Exec("INSERT INTO quiz_categories (name, grade, selected_books_json, total_questions, description) VALUES (?, ?, ?, ?, ?)",
		item.Name, item.Grade, string(booksJSON), item.TotalQuestions, item.Description)
	return err
}

func DeleteQuizCategory(id int64) error {
	_, err := db.DB.Exec("DELETE FROM quiz_categories WHERE id = ?", id)
	return err
}

func GetQuizCategoryByID(id int64) (*model.QuizCategoryItem, error) {
	var item model.QuizCategoryItem
	var booksJSON string
	err := db.DB.QueryRow("SELECT id, name, grade, selected_books_json, total_questions, description, created_at FROM quiz_categories WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Grade, &booksJSON, &item.TotalQuestions, &item.Description, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(booksJSON), &item.SelectedBooks)
	return &item, nil
}




