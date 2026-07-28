package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"coding_agent_web/internal/config"
	"coding_agent_web/internal/db"
	"coding_agent_web/internal/handler"
	"coding_agent_web/internal/knowledge"
)

func main() {
	config.LoadConfig()
	db.InitDB("/root/coding_agent_web/data.db")
	knowledge.LoadKnowledgeBase()

	// Public & Doc Routes
	http.HandleFunc("/", handler.HandleHome)
	http.HandleFunc("/api/doc/preview", handler.HandleDocPreview)

	// Auth Routes
	http.HandleFunc("/auth/login", handler.HandleAuthLoginView)
	http.HandleFunc("/auth/register", handler.HandleAuthRegisterView)
	http.HandleFunc("/api/auth/login", handler.HandleApiAuthLogin)
	http.HandleFunc("/api/auth/register", handler.HandleApiAuthRegister)
	http.HandleFunc("/api/auth/logout", handler.HandleApiAuthLogout)
	http.HandleFunc("/api/auth/me", handler.HandleApiAuthMe)

	// User Chat & History Routes
	http.HandleFunc("/api/chat", handler.HandleChat)
	http.HandleFunc("/api/sessions", handler.HandleGetSessions)
	http.HandleFunc("/api/session/messages", handler.HandleGetSessionMessages)
	http.HandleFunc("/api/session/delete", handler.HandleDeleteSession)
	http.HandleFunc("/api/quiz/generate", handler.HandleGenerateQuiz)
	http.HandleFunc("/api/quiz/score/save", handler.HandleSaveQuizScore)
	http.HandleFunc("/api/quiz/leaderboard", handler.HandleGetLeaderboard)

	// Admin Portal Web Page Routes (Separate URL Endpoints per Menu)
	http.HandleFunc("/admin", handler.HandleAdminUsersView)
	http.HandleFunc("/admin/users", handler.HandleAdminUsersView)
	http.HandleFunc("/admin/pdfs", handler.HandleAdminPDFsView)
	http.HandleFunc("/admin/quizzes", handler.HandleAdminQuizzesView)
	http.HandleFunc("/admin/packages", handler.HandleAdminPackagesView)
	http.HandleFunc("/admin/config", handler.HandleAdminConfigView)

	// Admin Portal Legacy Login (Optional)
	http.HandleFunc("/login", handler.HandleLogin)
	http.HandleFunc("/api/login", handler.HandleApiLogin)

	// Admin REST API Endpoints
	http.HandleFunc("/api/admin/config", handler.HandleAdminConfig)
	http.HandleFunc("/api/admin/users", handler.HandleAdminUsers)
	http.HandleFunc("/api/admin/user/delete", handler.HandleAdminUserDelete)
	http.HandleFunc("/api/admin/user/reset-password", handler.HandleAdminUserResetPassword)
	http.HandleFunc("/api/admin/user/set-limit", handler.HandleAdminUserSetLimit)

	http.HandleFunc("/api/admin/pdfs", handler.HandleAdminPDFs)
	http.HandleFunc("/api/admin/pdf/upload", handler.HandleAdminPDFUpload)
	http.HandleFunc("/api/admin/pdf/delete", handler.HandleAdminPDFDelete)

	http.HandleFunc("/api/admin/quizzes", handler.HandleAdminQuizzes)
	http.HandleFunc("/api/admin/quiz/create", handler.HandleAdminQuizCreate)
	http.HandleFunc("/api/admin/quiz/delete", handler.HandleAdminQuizDelete)

	// Quiz Category API Routes
	http.HandleFunc("/api/quiz/categories", handler.HandleGetQuizCategories)
	http.HandleFunc("/api/admin/quiz-categories", handler.HandleGetQuizCategories)
	http.HandleFunc("/api/admin/quiz-category/create", handler.HandleAdminQuizCategoryCreate)
	http.HandleFunc("/api/admin/quiz-category/delete", handler.HandleAdminQuizCategoryDelete)

	// Admin User Limit Route
	

	// Credit Package API Routes
	http.HandleFunc("/api/packages", handler.HandleGetCreditPackages)
	http.HandleFunc("/api/admin/packages", handler.HandleAdminCreditPackages)
	http.HandleFunc("/api/admin/package/create", handler.HandleAdminCreditPackageCreate)
	http.HandleFunc("/api/admin/package/update", handler.HandleAdminCreditPackageUpdate)
	http.HandleFunc("/api/admin/package/delete", handler.HandleAdminCreditPackageDelete)

	// Mayar.id Payment & Webhook Routes
	http.HandleFunc("/api/payment/transactions", handler.HandleGetUserPaymentTransactions)
	http.HandleFunc("/api/payment/cancel", handler.HandleCancelPaymentTransaction)
	http.HandleFunc("/api/payment/delete", handler.HandleDeletePaymentTransaction)
	http.HandleFunc("/api/payment/create-qris", handler.HandleCreateQRISTransaction)
	http.HandleFunc("/api/payment/status", handler.HandleGetPaymentStatus)
	http.HandleFunc("/api/mayar/webhook", handler.HandleMayarWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8097"
	}

	fmt.Printf("Clean Go Server starting on http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
