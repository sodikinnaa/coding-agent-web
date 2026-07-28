package model

import "time"

type AppConfig struct {
	AdminPassword string `json:"admin_password"`
	BaseURL       string `json:"base_url"`
	APIKey        string `json:"api_key"`
	Model         string `json:"model"`
	SystemPrompt  string `json:"system_prompt"`
	MayarAPIKey   string `json:"mayar_api_key"` // Mayar.id API Key
}

type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"`
	FullName       string    `json:"full_name"`
	Role           string    `json:"role"` // "admin" or "user"
	DailyLimit     int       `json:"daily_limit"`
	UsedToday      int       `json:"used_today"`
	RemainingToday int       `json:"remaining_today"`
	LastActiveDate string    `json:"last_active_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type ChatSession struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessageRecord struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type QuizScoreRecord struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	Grade          string    `json:"grade"`
	Score          int       `json:"score"`
	TotalQuestions int       `json:"total_questions"`
	CreatedAt      time.Time `json:"created_at"`
}

type LeaderboardEntry struct {
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	TotalScore int    `json:"total_score"`
	QuizCount  int    `json:"quiz_count"`
	HighGrade  string `json:"high_grade"`
}

type DocumentSnippet struct {
	BookName string `json:"book_name"`
	PageNum  int    `json:"page_num"`
	Content  string `json:"content"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []ContentPart
}

type OpenAIRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type DocItem struct {
	Filename  string `json:"filename"`
	RawName   string `json:"raw_name"`
	CharCount int    `json:"char_count"`
}

type AdminUserItem struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	Role       string    `json:"role"`
	DailyLimit int       `json:"daily_limit"`
	UsedToday  int       `json:"used_today"`
	CreatedAt  time.Time `json:"created_at"`
	QuizCount  int       `json:"quiz_count"`
	TotalScore int       `json:"total_score"`
}

type AdminPDFItem struct {
	Filename  string `json:"filename"`
	RawName   string `json:"raw_name"`
	SizeBytes int64  `json:"size_bytes"`
	ModTime   string `json:"mod_time"`
}

type CustomQuizItem struct {
	ID            int64     `json:"id"`
	Grade         string    `json:"grade"`
	Topic         string    `json:"topic"`
	Question      string    `json:"question"`
	Options       []string  `json:"options"`
	CorrectIndex  int       `json:"correct_index"`
	Explanation   string    `json:"explanation"`
	ReferenceBook string    `json:"reference_book"`
	CreatedAt     time.Time `json:"created_at"`
}

type QuizCategoryItem struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Grade          string    `json:"grade"`
	SelectedBooks  []string  `json:"selected_books"`
	TotalQuestions int       `json:"total_questions"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaymentTransaction struct {
	ID         string     `json:"id"`
	UserID     int64      `json:"user_id"`
	TierName   string     `json:"tier_name"`
	DailyLimit int        `json:"daily_limit"`
	Amount     int        `json:"amount"`
	Status     string     `json:"status"` // "pending", "paid", "expired"
	QRURL      string     `json:"qr_url"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiredAt  *time.Time `json:"expired_at,omitempty"`
}

type CreditPackageItem struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	DailyLimit  int       `json:"daily_limit"`
	Price       int       `json:"price"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
