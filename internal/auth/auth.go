package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"coding_agent_web/internal/config"
	"coding_agent_web/internal/db"
	"coding_agent_web/internal/model"

	"golang.org/x/crypto/bcrypt"
)

var (
	sessions     = make(map[string]int64) // token -> user_id
	sessionMutex sync.RWMutex
)

func EnsureAdminUserExists() {
	cfg := config.GetConfig()
	adminPass := cfg.AdminPassword
	if adminPass == "" {
		adminPass = "L4njutk4n"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning: Failed to bcrypt admin password: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")
	var count int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)

	if count == 0 {
		_, err = db.DB.Exec(`
			INSERT INTO users (username, password_hash, full_name, role, daily_limit, used_today, last_active_date)
			VALUES ('admin', ?, 'System Administrator', 'admin', 999999, 0, ?)
		`, string(hash), today)
		if err != nil {
			log.Printf("Warning: Failed to seed admin user: %v", err)
		} else {
			log.Println("✅ Initialized default admin account 'admin' into SQLite database.")
		}
	} else {
		_, _ = db.DB.Exec("UPDATE users SET password_hash = ? WHERE username = 'admin'", string(hash))
		log.Println("✅ Synced admin account 'admin' password hash in SQLite database.")
	}
}

func RegisterUser(username, password, fullName string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("Username dan password tidak boleh kosong")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if fullName == "" {
		fullName = username
	}

	role := "user"
	dailyLimit := 5 // Default free tier: 5 chats / day
	if username == "admin" {
		role = "admin"
		dailyLimit = 999999
	}

	res, err := db.DB.Exec("INSERT INTO users (username, password_hash, full_name, role, daily_limit, used_today, last_active_date) VALUES (?, ?, ?, ?, ?, 0, ?)", username, string(hash), fullName, role, dailyLimit, time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, errors.New("Username sudah terdaftar atau terjadi kesalahan")
	}

	id, _ := res.LastInsertId()
	return &model.User{
		ID:             id,
		Username:       username,
		FullName:       fullName,
		Role:           role,
		DailyLimit:     dailyLimit,
		UsedToday:      0,
		RemainingToday: dailyLimit,
		LastActiveDate: time.Now().Format("2006-01-02"),
	}, nil
}

func LoginUser(username, password string) (string, *model.User, error) {
	var user model.User
	var hash string

	row := db.DB.QueryRow("SELECT id, username, password_hash, full_name, role, daily_limit, used_today, last_active_date, created_at FROM users WHERE username = ?", username)
	err := row.Scan(&user.ID, &user.Username, &hash, &user.FullName, &user.Role, &user.DailyLimit, &user.UsedToday, &user.LastActiveDate, &user.CreatedAt)
	if err != nil {
		return "", nil, errors.New("Username atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", nil, errors.New("Username atau password salah")
	}

	// Auto Reset Daily Usage if date changed
	today := time.Now().Format("2006-01-02")
	if user.LastActiveDate != today {
		user.UsedToday = 0
		user.LastActiveDate = today
		db.DB.Exec("UPDATE users SET used_today = 0, last_active_date = ? WHERE id = ?", today, user.ID)
	}

	user.RemainingToday = user.DailyLimit - user.UsedToday
	if user.RemainingToday < 0 {
		user.RemainingToday = 0
	}

	token := generateToken()
	sessionMutex.Lock()
	sessions[token] = user.ID
	sessionMutex.Unlock()

	return token, &user, nil
}

func LogoutUser(token string) {
	sessionMutex.Lock()
	delete(sessions, token)
	sessionMutex.Unlock()
}

func GetUserByToken(token string) (*model.User, error) {
	sessionMutex.RLock()
	userID, exists := sessions[token]
	sessionMutex.RUnlock()

	if !exists {
		return nil, errors.New("Unauthorized")
	}

	var user model.User
	row := db.DB.QueryRow("SELECT id, username, full_name, role, daily_limit, used_today, last_active_date, created_at FROM users WHERE id = ?", userID)
	err := row.Scan(&user.ID, &user.Username, &user.FullName, &user.Role, &user.DailyLimit, &user.UsedToday, &user.LastActiveDate, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	if user.LastActiveDate != today {
		user.UsedToday = 0
		user.LastActiveDate = today
		db.DB.Exec("UPDATE users SET used_today = 0, last_active_date = ? WHERE id = ?", today, user.ID)
	}

	user.RemainingToday = user.DailyLimit - user.UsedToday
	if user.RemainingToday < 0 {
		user.RemainingToday = 0
	}

	return &user, nil
}

func GetUserFromRequest(r *http.Request) (*model.User, error) {
	cookie, err := r.Cookie("user_token")
	if err == nil && cookie.Value != "" {
		if u, err := GetUserByToken(cookie.Value); err == nil {
			return u, nil
		}
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return GetUserByToken(authHeader[7:])
	}

	return nil, errors.New("Unauthorized")
}

func CheckAndUpdateDailyUsage(userID int64) (int, int, error) {
	var user model.User
	row := db.DB.QueryRow("SELECT id, role, daily_limit, used_today, last_active_date FROM users WHERE id = ?", userID)
	err := row.Scan(&user.ID, &user.Role, &user.DailyLimit, &user.UsedToday, &user.LastActiveDate)
	if err != nil {
		return 0, 0, err
	}

	if user.Role == "admin" {
		return 999999, 999999, nil
	}

	today := time.Now().Format("2006-01-02")
	if user.LastActiveDate != today {
		user.UsedToday = 0
		user.LastActiveDate = today
		db.DB.Exec("UPDATE users SET used_today = 0, last_active_date = ? WHERE id = ?", today, userID)
	}

	if user.UsedToday >= user.DailyLimit {
		return 0, user.DailyLimit, errors.New("Batas kredit chat harian kamu sudah habis. Silakan pilih paket kredit yang lebih tinggi atau tunggu reset esok hari!")
	}

	user.UsedToday++
	db.DB.Exec("UPDATE users SET used_today = ?, last_active_date = ? WHERE id = ?", user.UsedToday, today, userID)
	remaining := user.DailyLimit - user.UsedToday
	if remaining < 0 {
		remaining = 0
	}
	return remaining, user.DailyLimit, nil
}

func generateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + fmt.Sprintf("%d", time.Now().UnixNano())
}
