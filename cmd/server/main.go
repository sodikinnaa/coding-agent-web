package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"coding_agent_web/internal/config"
	"coding_agent_web/internal/handler"
	"coding_agent_web/internal/knowledge"
)

func main() {
	config.LoadConfig()
	knowledge.LoadKnowledgeBase()

	http.HandleFunc("/", handler.HandleHome)
	http.HandleFunc("/api/chat", handler.HandleChat)
	http.HandleFunc("/api/quiz/generate", handler.HandleGenerateQuiz)
	http.HandleFunc("/api/quiz/score/save", handler.HandleSaveQuizScore)
	http.HandleFunc("/api/quiz/leaderboard", handler.HandleGetLeaderboard)
	http.HandleFunc("/login", handler.HandleLogin)
	http.HandleFunc("/api/login", handler.HandleApiLogin)
	http.HandleFunc("/admin", handler.HandleAdminUsersView)
	http.HandleFunc("/admin/users", handler.HandleAdminUsersView)
	http.HandleFunc("/admin/pdfs", handler.HandleAdminPDFsView)
	http.HandleFunc("/admin/quizzes", handler.HandleAdminQuizzesView)
	http.HandleFunc("/admin/config", handler.HandleAdminConfigView)
	http.HandleFunc("/api/admin/config", handler.HandleAdminConfig)
	http.HandleFunc("/api/admin/users", handler.HandleAdminUsers)
	http.HandleFunc("/api/admin/user/delete", handler.HandleAdminUserDelete)
	http.HandleFunc("/api/admin/user/reset-password", handler.HandleAdminUserResetPassword)
	http.HandleFunc("/api/admin/user/set-limit", handler.HandleAdminUserSetLimit)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8097"
	}

	fmt.Printf("Clean Go Server starting on http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
