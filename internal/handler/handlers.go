package handler

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"coding_agent_web/internal/auth"
	"coding_agent_web/internal/config"
	"coding_agent_web/internal/knowledge"
	"coding_agent_web/internal/model"
	"coding_agent_web/internal/service"
)

type QuizQuestion struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectIndex  int      `json:"correct_index"`
	Explanation   string   `json:"explanation"`
	ReferenceBook string   `json:"reference_book"`
}

// Auth API Handlers
func HandleAuthLoginView(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleAuthRegisterView(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleApiAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Format request tidak valid"})
		return
	}

	user, err := auth.RegisterUser(req.Username, req.Password, req.FullName)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	// Auto login on successful registration
	token, _, err := auth.LoginUser(req.Username, req.Password)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "user_token",
			Value:    token,
			Path:     "/",
			MaxAge:   86400 * 30, // 30 Days Persistent Session
			HttpOnly: true,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
		"token":   token,
	})
}

func HandleApiAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"remember_me"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Format request tidak valid"})
		return
	}

	token, user, err := auth.LoginUser(req.Username, req.Password)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	var maxAge = 86400 * 30 // 30 Days default session (Remember Me)
	if req.RememberMe {
		maxAge = 86400 * 365 // 1 Year session
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"user":      user,
		"token":     token,
		"redirect":  user.Role == "admin",
	})
}

func HandleApiAuthLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_token")
	if err == nil && cookie.Value != "" {
		auth.LogoutUser(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "user_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleApiAuthMe(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"authenticated": false})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user":          user,
	})
}

// User Chat Sessions & History Handlers
func HandleGetSessions(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	sessions, err := service.GetUserSessions(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if sessions == nil {
		sessions = []model.ChatSession{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"sessions": sessions})
}

func HandleGetSessionMessages(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing session id"})
		return
	}

	messages, err := service.GetSessionMessages(sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if messages == nil {
		messages = []model.ChatMessageRecord{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"messages": messages})
}

func HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		req.SessionID = r.URL.Query().Get("id")
	}

	if req.SessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing session id"})
		return
	}

	if err := service.DeleteChatSession(user.ID, req.SessionID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Kurikulum Koding & AI SD-SMA</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
        body { font-family: 'Inter', sans-serif; }
        .assistant-ui-root { font-family: 'Inter', sans-serif; }
        .assistant-ui-thread { max-width: 48rem; margin: 0 auto; width: 100%; }
        .aui-msg-user { background-color: #27272a; color: #f4f4f5; border-radius: 1.25rem 1.25rem 0.25rem 1.25rem; border: 1px solid #3f3f46; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
        .aui-msg-assistant { background-color: #18181b; color: #e4e4e7; border: 1px solid #27272a; border-radius: 1.25rem 1.25rem 1.25rem 0.25rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
        .aui-pricing-card { background-color: #18181b; border: 1px solid #27272a; border-radius: 1rem; transition: all 0.2s ease-in-out; }
        .aui-pricing-card:hover { border-color: #3b82f6; transform: translateY(-2px); }
        .aui-pricing-card.featured { border-color: #3b82f6; background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%); box-shadow: 0 10px 25px -5px rgba(59, 130, 246, 0.2); }
        .markdown-body p { margin-bottom: 0.75rem; }
        .markdown-body p:last-child { margin-bottom: 0; }
        .markdown-body ul, .markdown-body ol { margin-left: 1.25rem; margin-bottom: 0.75rem; }
        .markdown-body ul { list-style-type: disc; }
        .markdown-body ol { list-style-type: decimal; }
        .markdown-body li { margin-bottom: 0.25rem; }
        .markdown-body strong { color: #f4f4f5; font-weight: 600; }
        .markdown-body em { color: #93c5fd; font-style: italic; }
        .markdown-body code { background: #27272a; padding: 0.15rem 0.4rem; border-radius: 0.25rem; font-size: 0.85em; font-family: monospace; color: #60a5fa; }
        .markdown-body hr { border-color: #27272a; margin: 1rem 0; }
        .markdown-body h1, .markdown-body h2, .markdown-body h3 { font-weight: 600; color: #f4f4f5; margin-top: 1rem; margin-bottom: 0.5rem; }
        .markdown-body h1 { font-size: 1.25rem; }
        .markdown-body h2 { font-size: 1.1rem; }
        .markdown-body h3 { font-size: 1rem; }
    </style>
</head>
<body class="bg-black text-zinc-100 flex flex-col h-screen overflow-hidden">
    <!-- Navbar -->
    <header class="h-14 border-b border-zinc-800 bg-zinc-950/80 backdrop-blur px-3 md:px-6 flex justify-between items-center z-20 shrink-0">
        <div class="flex items-center space-x-2.5">
            <!-- Mobile Sidebar Toggle Button -->
            <button onclick="toggleMobileSidebar()" class="lg:hidden text-zinc-400 hover:text-white p-2 rounded-lg bg-zinc-900 border border-zinc-800 transition active:scale-95">
                <i class="fa-solid fa-book-open text-xs"></i>
            </button>

            <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center text-white font-semibold text-sm shadow-lg shadow-blue-500/20 shrink-0">
                <i class="fa-solid fa-sparkles text-xs"></i>
            </div>
            <!-- Navbar Badge -->
            <div>
                <h1 class="text-xs md:text-sm font-semibold text-zinc-100 flex items-center gap-1.5">
                    <span class="truncate max-w-[120px] sm:max-w-none">Kurikulum Koding & AI</span>
                    <span class="hidden sm:inline-block px-2 py-0.5 text-[10px] font-medium bg-gradient-to-r from-blue-500/10 to-indigo-500/10 text-blue-400 border border-blue-500/20 rounded-full flex items-center gap-1">
                        <i class="fa-solid fa-eye text-[9px]"></i> Direct PDF Vision AI & Quiz
                    </span>
                </h1>
            </div>
        </div>

        <!-- Toggle Mode: Chat vs Quiz -->
        <div class="flex items-center bg-zinc-900 border border-zinc-800 p-0.5 md:p-1 rounded-xl">
            <button id="btn-mode-chat" onclick="switchMode('chat')" class="px-2.5 md:px-3 py-1 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center gap-1.5 shadow-sm">
                <i class="fa-regular fa-comments"></i>
                <span class="hidden sm:inline">Tanya Jawab AI</span>
                <span class="sm:hidden">Chat</span>
            </button>
            <button id="btn-mode-quiz" onclick="switchMode('quiz')" class="px-2.5 md:px-3 py-1 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center gap-1.5">
                <i class="fa-solid fa-list-check"></i>
                <span class="hidden sm:inline">Kuis Interaktif</span>
                <span class="sm:hidden">Kuis</span>
            </button>
        </div>

        <div class="flex items-center space-x-2">
            <!-- User Auth Container -->
            <div id="auth-container" class="flex items-center space-x-2">
                <button onclick="openAuthModal('login')" class="text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white px-3 py-1.5 rounded-lg transition shadow-sm flex items-center gap-1.5">
                    <i class="fa-solid fa-right-to-bracket text-[10px]"></i>
                    <span>Masuk</span>
                </button>
                <button onclick="openAuthModal('register')" class="text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 px-3 py-1.5 rounded-lg border border-zinc-700 transition flex items-center gap-1.5 shadow-sm">
                    <i class="fa-solid fa-user-plus text-[10px]"></i>
                    <span>Daftar</span>
                </button>
            </div>

            <a href="/login" class="text-xs font-medium bg-zinc-900 hover:bg-zinc-800 text-zinc-400 hover:text-zinc-200 p-2 rounded-lg border border-zinc-800 transition flex items-center gap-1.5 shadow-sm" title="Admin Portal">
                <i class="fa-solid fa-lock text-xs"></i>
            </a>
        </div>
    </header>

    <!-- Main Container -->
    <div class="flex-1 flex overflow-hidden relative">
        <!-- Sidebar Docs & History (Desktop & Mobile Drawer) -->
        <div id="mobile-sidebar-backdrop" onclick="toggleMobileSidebar()" class="fixed inset-0 bg-black/70 backdrop-blur-sm z-30 hidden lg:hidden transition-opacity"></div>

        <aside id="sidebar-docs" class="w-72 md:w-80 bg-zinc-950 border-r border-zinc-800 p-4 fixed lg:static inset-y-0 left-0 z-40 transform -translate-x-full lg:translate-x-0 transition-transform duration-300 ease-in-out flex flex-col h-full">
            
            <!-- Sidebar Navigation Tabs: History vs Knowledge Base -->
            <div class="flex items-center bg-zinc-900 border border-zinc-800 p-1 rounded-xl mb-4 shrink-0">
                <button id="tab-btn-history" onclick="switchSidebarTab('history')" class="flex-1 py-1.5 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center justify-center gap-1.5">
                    <i class="fa-solid fa-clock-rotate-left text-[11px]"></i>
                    <span>Riwayat Chat</span>
                </button>
                <button id="tab-btn-docs" onclick="switchSidebarTab('docs')" class="flex-1 py-1.5 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5">
                    <i class="fa-solid fa-books text-[11px]"></i>
                    <span>Knowledge Base</span>
                </button>
            </div>

            <!-- Tab 1: User Chat History (Default Active) -->
            <div id="tab-content-history" class="flex-1 flex flex-col min-h-0">
                <div class="mb-3 px-1 shrink-0 space-y-2">
                    <button onclick="startNewChatSession()" class="w-full py-2.5 px-3 bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-zinc-100 font-semibold text-xs rounded-xl transition flex items-center justify-between group shadow-sm active:scale-98">
                        <div class="flex items-center gap-2">
                            <div class="w-5 h-5 rounded-lg bg-blue-600/20 text-blue-400 border border-blue-500/30 flex items-center justify-center text-[10px]">
                                <i class="fa-solid fa-plus"></i>
                            </div>
                            <span>Percakapan Baru</span>
                        </div>
                        <i class="fa-regular fa-pen-to-square text-zinc-500 group-hover:text-zinc-300 text-xs"></i>
                    </button>
                    <div class="flex items-center justify-between pt-1">
                        <h2 class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Riwayat Obrolan</h2>
                        <span id="session-count-badge" class="text-[10px] bg-zinc-900 text-zinc-400 px-2 py-0.5 rounded-full font-mono border border-zinc-800">0 Chat</span>
                    </div>
                </div>
                
                <!-- History Login Guard Notice -->
                <div id="history-login-notice" class="hidden p-4 bg-zinc-900/80 border border-zinc-800 rounded-2xl text-center space-y-2.5 my-auto shadow-xl">
                    <div class="w-10 h-10 rounded-xl bg-blue-600/10 border border-blue-500/20 text-blue-400 flex items-center justify-center text-lg mx-auto">
                        <i class="fa-solid fa-user-lock"></i>
                    </div>
                    <p class="text-xs text-zinc-200 font-bold">Simpan Riwayat Percakapan</p>
                    <p class="text-[11px] text-zinc-400 leading-relaxed">Masuk akun untuk menyimpan dan mengakses semua percakapan AI Anda kapan saja.</p>
                    <button onclick="openAuthModal('login')" class="w-full mt-1 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold py-2 rounded-xl transition shadow-md">
                        Masuk / Daftar Akun
                    </button>
                </div>

                <div id="history-sessions-list" class="flex-1 overflow-y-auto space-y-1.5 pr-1">
                    <!-- Dynamic sessions list -->
                </div>
            </div>

            <!-- Tab 2: Knowledge Base Docs -->
            <div id="tab-content-docs" class="flex-1 flex flex-col min-h-0 hidden">
                <div class="flex items-center justify-between mb-3 px-1 shrink-0">
                    <h2 class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">Dokumen Referensi</h2>
                    <span class="text-[10px] bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded-full font-mono">{{len .Docs}} Ref</span>
                </div>
                <div class="flex-1 overflow-y-auto space-y-2 pr-1">
                    {{range .Docs}}
                    <div class="p-3 bg-zinc-900/60 hover:bg-zinc-900 rounded-xl border border-zinc-800/80 transition group flex flex-col gap-2 shadow-sm">
                        <div onclick="previewDoc('{{.RawName}}', '{{.Filename}}', 'pdf')" class="cursor-pointer">
                            <div class="flex items-center space-x-2 text-zinc-200 group-hover:text-blue-400 text-xs font-medium mb-1">
                                <i class="fa-regular fa-bookmark text-blue-400 shrink-0"></i>
                                <span class="truncate leading-snug">{{.Filename}}</span>
                            </div>
                            <p class="text-[10px] text-zinc-500 font-mono">{{.CharCount}} bytes</p>
                        </div>
                        <div class="flex items-center justify-end space-x-1.5 shrink-0 pt-1 border-t border-zinc-800/50">
                            <button onclick="previewDoc('{{.RawName}}', '{{.Filename}}', 'pdf')" title="Preview PDF Visual" class="w-full py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-[10px] font-medium transition flex items-center justify-center gap-1.5 active:scale-95">
                                <i class="fa-solid fa-file-pdf"></i>
                                <span>Preview PDF / DOCX</span>
                            </button>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
                <div class="mb-3 px-1 shrink-0 space-y-2">
                    <button onclick="startNewChatSession()" class="w-full py-2.5 px-3 bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-zinc-100 font-semibold text-xs rounded-xl transition flex items-center justify-between group shadow-sm active:scale-98">
                        <div class="flex items-center gap-2">
                            <div class="w-5 h-5 rounded-lg bg-blue-600/20 text-blue-400 border border-blue-500/30 flex items-center justify-center text-[10px]">
                                <i class="fa-solid fa-plus"></i>
                            </div>
                            <span>Percakapan Baru</span>
                        </div>
                        <i class="fa-regular fa-pen-to-square text-zinc-500 group-hover:text-zinc-300 text-xs"></i>
                    </button>
                    <div class="flex items-center justify-between pt-1">
                        <h2 class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Riwayat Obrolan</h2>
                        <span id="session-count-badge" class="text-[10px] bg-zinc-900 text-zinc-400 px-2 py-0.5 rounded-full font-mono border border-zinc-800">0 Chat</span>
                    </div>
                </div>
                
                <!-- History Login Guard Notice -->
                <div id="history-login-notice" class="hidden p-4 bg-zinc-900/80 border border-zinc-800 rounded-2xl text-center space-y-2.5 my-auto shadow-xl">
                    <div class="w-10 h-10 rounded-xl bg-blue-600/10 border border-blue-500/20 text-blue-400 flex items-center justify-center text-lg mx-auto">
                        <i class="fa-solid fa-user-lock"></i>
                    </div>
                    <p class="text-xs text-zinc-200 font-bold">Simpan Riwayat Percakapan</p>
                    <p class="text-[11px] text-zinc-400 leading-relaxed">Masuk akun untuk menyimpan dan mengakses semua percakapan AI Anda kapan saja.</p>
                    <button onclick="openAuthModal('login')" class="w-full mt-1 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold py-2 rounded-xl transition shadow-md">
                        Masuk / Daftar Akun
                    </button>
                </div>

                <div id="history-sessions-list" class="flex-1 overflow-y-auto space-y-1.5 pr-1">
                    <!-- Dynamic sessions list -->
                </div>
            </div>

            <!-- ChatGPT-style Sidebar User Account Footer -->
            <div id="sidebar-user-footer" class="mt-auto pt-3 border-t border-zinc-800/80 shrink-0 hidden space-y-2">
                <div onclick="openProfileModal()" class="p-2.5 bg-zinc-900/90 hover:bg-zinc-800/90 border border-zinc-800 hover:border-zinc-700 rounded-2xl cursor-pointer transition shadow-lg flex items-center justify-between group">
                    <div class="flex items-center space-x-2.5 min-w-0">
                        <div id="sb-user-avatar" class="w-8 h-8 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center text-xs font-bold shrink-0 shadow-md">U</div>
                        <div class="min-w-0">
                            <div id="sb-user-name" class="text-xs font-semibold text-zinc-100 truncate leading-snug">User</div>
                            <div class="flex items-center gap-1.5 mt-0.5">
                                <span id="sb-user-plan-badge" class="px-1.5 py-0.2 text-[9px] font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-md">Free Plan</span>
                                <span id="sb-user-usage-str" class="text-[10px] text-zinc-400 font-mono">0/5</span>
                            </div>
                        </div>
                    </div>
                    <i class="fa-solid fa-gear text-zinc-400 group-hover:text-zinc-200 text-xs px-1"></i>
                </div>
                
                <button onclick="openProfileModal(); switchProfTab('billing');" class="w-full py-2 px-3 bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-blue-500/20 transition flex items-center justify-center gap-2 active:scale-98">
                    <i class="fa-solid fa-sparkles text-amber-300 text-[11px]"></i>
                    <span>Langganan & Billing</span>
                </button>
            </div>

        </aside>

        <!-- VIEW 1: CHATBOT INTERFACE -->
        <main id="chat-view" class="flex-1 flex flex-col bg-zinc-950 relative">
            <div id="chat-box" class="flex-1 overflow-y-auto p-4 md:p-6 space-y-6 assistant-ui-thread">
                <div class="flex items-start gap-3">
                    <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xs shrink-0 shadow-md">
                        <i class="fa-solid fa-sparkles"></i>
                    </div>
                    <div class="aui-msg-assistant p-4 text-sm text-zinc-200 leading-relaxed shadow-sm space-y-2">
                        <p>Halo! Saya Asisten AI Kurikulum Koding & AI SD-SMA.</p>
                        <div class="p-2.5 rounded-xl bg-blue-500/10 border border-blue-500/20 text-xs text-blue-300 flex items-center gap-2">
                            <i class="fa-solid fa-wand-magic-sparkles text-blue-400"></i>
                            <span>Didukung oleh <strong>Direct PDF Vision AI</strong> (Analisis Berkas PDF Visual secara Presisi).</span>
                        </div>
                        <p class="text-xs text-zinc-400 pt-1">Silakan tanyakan materi kurikulum, atau beralih ke <strong>Kuis Interaktif</strong> di tombol atas!</p>
                    </div>
                </div>
            </div>

            <div class="p-4 bg-gradient-to-t from-zinc-950 via-zinc-950 to-transparent">
                <!-- Locked Chat Overlay for Guest -->
                <div id="chat-guest-lock" class="hidden p-4 bg-zinc-900/90 border border-zinc-800 backdrop-blur-md rounded-2xl text-center space-y-2.5 shadow-2xl">
                    <div class="flex items-center justify-center gap-2 text-amber-400 font-bold text-xs">
                        <i class="fa-solid fa-lock text-sm"></i>
                        <span>Obrolan AI Terkunci</span>
                    </div>
                    <p class="text-xs text-zinc-300">Harap masuk atau daftar akun terlebih dahulu untuk melakukan obrolan dengan AI Kurikulum.</p>
                    <div class="flex justify-center gap-2.5 pt-1">
                        <button onclick="openAuthModal('login')" class="bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold px-4 py-2 rounded-xl transition shadow-md flex items-center gap-1.5">
                            <i class="fa-solid fa-right-to-bracket text-[10px]"></i> Masuk Sekarang
                        </button>
                        <button onclick="openAuthModal('register')" class="bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-semibold px-4 py-2 rounded-xl border border-zinc-700 transition flex items-center gap-1.5">
                            <i class="fa-solid fa-user-plus text-[10px]"></i> Daftar Akun
                        </button>
                    </div>
                </div>

                <div id="chat-input-wrapper" class="assistant-ui-thread">
                    <form id="chat-form" class="relative flex items-center bg-zinc-900 border border-zinc-800 focus-within:border-zinc-700 rounded-2xl shadow-xl transition-all p-1.5">
                        <textarea id="user-input" rows="1" placeholder="Tanyakan seputar kurikulum koding SD-SMA..." class="w-full bg-transparent text-zinc-100 placeholder-zinc-500 px-4 py-2.5 text-sm focus:outline-none resize-none max-h-32"></textarea>
                        <button type="submit" class="bg-blue-600 hover:bg-blue-500 text-white w-9 h-9 rounded-xl flex items-center justify-center transition shrink-0 ml-2 shadow-md">
                            <i class="fa-solid fa-arrow-up text-sm"></i>
                        </button>
                    </form>
                    <div class="flex items-center justify-between text-[11px] text-zinc-500 mt-2 px-1">
                        <span>Powered by Assistant-UI & Go RAG Engine</span>
                        <span id="chat-credit-indicator" class="font-medium text-blue-400"></span>
                    </div>
                </div>
            </div>
        </main>

        <!-- VIEW 2: INTERACTIVE QUIZ INTERFACE -->
        <main id="quiz-view" class="flex-1 flex-col bg-zinc-950 p-6 overflow-y-auto hidden">
            <div class="max-w-2xl mx-auto w-full space-y-6">
                <!-- Guest Auth Banner -->
                <div id="quiz-auth-banner" class="hidden p-6 bg-zinc-900 border border-zinc-800 rounded-2xl text-center space-y-3 shadow-xl">
                    <div class="w-12 h-12 rounded-2xl bg-blue-600/10 border border-blue-500/20 text-blue-400 flex items-center justify-center text-xl mx-auto">
                        <i class="fa-solid fa-lock"></i>
                    </div>
                    <h3 class="text-sm font-bold text-zinc-100">Harap Login untuk Mengikuti Kuis & Papan Peringkat</h3>
                    <p class="text-xs text-zinc-400">Silakan masuk atau daftar akun terlebih dahulu untuk mengakses fitur kuis interaktif dan mencatatkan skor Anda di papan peringkat.</p>
                    <div class="flex justify-center gap-3 pt-2">
                        <button onclick="openAuthModal('login')" class="bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold px-4 py-2 rounded-xl transition">
                            Masuk
                        </button>
                        <button onclick="openAuthModal('register')" class="bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-semibold px-4 py-2 rounded-xl border border-zinc-700 transition">
                            Daftar
                        </button>
                    </div>
                </div>

                <!-- Quiz Main Content (visible when logged in) -->
                <div id="quiz-main-content" class="space-y-6">
                    <!-- Selector Kategori Kuis -->
                    <div class="bg-zinc-900 border border-zinc-800 rounded-2xl p-5 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <h2 class="text-base font-bold text-white flex items-center gap-2">
                                <i class="fa-solid fa-graduation-cap text-blue-500"></i>
                                <span>Pilih Kategori Kuis Koding & AI</span>
                            </h2>
                            <p class="text-xs text-zinc-400 mt-0.5">Kategori & Buku Referensi di-set oleh Admin, soal di-generate dinamis oleh AI</p>
                        </div>
                        <div class="flex items-center space-x-3">
                            <select id="quiz-category-select" class="bg-zinc-950 border border-zinc-800 text-white text-xs rounded-xl px-4 py-2.5 focus:outline-none focus:border-blue-500 min-w-[200px]">
                                <option value="0">Pilih Kategori Kuis...</option>
                            </select>
                            <button onclick="generateQuiz()" class="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white px-4 py-2.5 rounded-xl text-xs font-semibold shadow-lg shadow-blue-500/20 transition flex items-center gap-2 shrink-0">
                                <i class="fa-solid fa-wand-magic-sparkles"></i>
                                <span>Mulai Kuis</span>
                            </button>
                        </div>
                    </div>

                    <!-- Quiz Container -->
                    <div id="quiz-container" class="hidden space-y-6">
                        <!-- Progress Bar -->
                        <div class="flex items-center justify-between text-xs text-zinc-400 mb-1">
                            <span id="quiz-progress-text">Soal 1 dari 5</span>
                            <span id="quiz-score-live" class="font-semibold text-blue-400">Skor: 0</span>
                        </div>
                        <div class="w-full bg-zinc-900 h-2 rounded-full overflow-hidden border border-zinc-800">
                            <div id="quiz-progress-bar" class="bg-blue-600 h-full transition-all duration-300" style="width: 20%;"></div>
                        </div>

                        <!-- Question Card -->
                        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl p-6 shadow-xl space-y-5">
                            <div class="space-y-2">
                                <span id="quiz-ref-tag" class="inline-block px-2.5 py-1 text-[10px] font-mono bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-md"></span>
                                <h3 id="quiz-question-text" class="text-base font-semibold text-zinc-100 leading-snug"></h3>
                            </div>

                            <!-- Options List -->
                            <div id="quiz-options-box" class="space-y-2.5">
                                <!-- Options dynamically rendered -->
                            </div>

                            <!-- Explanation Box (Initially Hidden) -->
                            <div id="quiz-explanation" class="hidden p-4 bg-zinc-950 border border-zinc-800 rounded-xl space-y-2 text-xs">
                                <div id="quiz-result-badge" class="font-bold"></div>
                                <p id="quiz-explanation-text" class="text-zinc-300 leading-relaxed"></p>
                            </div>
                        </div>

                        <!-- Navigation Buttons -->
                        <div class="flex justify-between items-center pt-2">
                            <button id="btn-prev-q" onclick="prevQuestion()" class="hidden text-xs bg-zinc-900 hover:bg-zinc-800 text-zinc-300 px-4 py-2.5 rounded-xl border border-zinc-800 transition items-center gap-1.5">
                                <i class="fa-solid fa-arrow-left"></i>
                                <span>Sebelumnya</span>
                            </button>
                            <button id="btn-next-q" onclick="nextQuestion()" class="ml-auto text-xs bg-blue-600 hover:bg-blue-500 text-white font-semibold px-5 py-2.5 rounded-xl transition flex items-center gap-1.5 shadow-md">
                                <span>Soal Berikutnya</span>
                                <i class="fa-solid fa-arrow-right"></i>
                            </button>
                        </div>
                    </div>

                    <!-- Quiz Score Result Modal (Final Score) -->
                    <div id="quiz-final-result" class="hidden bg-zinc-900 border border-zinc-800 rounded-2xl p-8 text-center space-y-4 shadow-2xl">
                        <div class="w-16 h-16 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center text-2xl mx-auto shadow-lg shadow-blue-500/20">
                            <i class="fa-solid fa-trophy"></i>
                        </div>
                        <h3 class="text-xl font-bold text-white">Kuis Selesai!</h3>
                        <p class="text-xs text-zinc-400">Berikut adalah hasil capaian kuis koding kamu:</p>

                        <div class="text-4xl font-extrabold text-blue-400 py-2 font-mono" id="final-score-val">100 / 100</div>

                        <button onclick="generateQuiz()" class="bg-blue-600 hover:bg-blue-500 text-white px-6 py-2.5 rounded-xl text-xs font-semibold transition inline-flex items-center gap-2">
                            <i class="fa-solid fa-rotate-right"></i>
                            <span>Coba Kuis Lagi</span>
                        </button>
                    </div>

                    <!-- Loading Spinner for Quiz -->
                    <div id="quiz-loading" class="hidden p-12 text-center space-y-3">
                        <i class="fa-solid fa-spinner animate-spin text-3xl text-blue-500"></i>
                        <p class="text-xs text-zinc-400">Sedang menyusun soal kuis berdasarkan dokumen kurikulum...</p>
                    </div>
                </div>

                <!-- Leaderboard Table Component -->
                <div class="bg-zinc-900 border border-zinc-800 rounded-2xl p-6 shadow-xl space-y-4">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center space-x-3">
                            <div class="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center text-base">
                                <i class="fa-solid fa-trophy"></i>
                            </div>
                            <div>
                                <h3 class="text-sm font-bold text-zinc-100">Papan Peringkat (Leaderboard)</h3>
                                <p class="text-xs text-zinc-400">Pemain terbaik berdasarkan akumulasi skor kuis</p>
                            </div>
                        </div>
                        <button onclick="loadLeaderboard()" class="text-xs text-zinc-400 hover:text-white p-2 rounded-lg bg-zinc-950 border border-zinc-800 transition">
                            <i class="fa-solid fa-rotate"></i>
                        </button>
                    </div>

                    <div class="overflow-x-auto">
                        <table class="w-full text-left text-xs">
                            <thead>
                                <tr class="border-b border-zinc-800 text-zinc-500 font-semibold uppercase tracking-wider">
                                    <th class="py-2.5 px-3">Peringkat</th>
                                    <th class="py-2.5 px-3">Pemain</th>
                                    <th class="py-2.5 px-3 text-center">Jenjang</th>
                                    <th class="py-2.5 px-3 text-center">Jumlah Kuis</th>
                                    <th class="py-2.5 px-3 text-right">Total Skor</th>
                                </tr>
                            </thead>
                            <tbody id="leaderboard-tbody" class="divide-y divide-zinc-800/50">
                                <!-- Dynamic Leaderboard Rows -->
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <!-- Auth Modal (Login & Register) -->
    <div id="auth-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-md p-6 shadow-2xl space-y-5">
            <div class="flex items-center justify-between">
                <div class="flex items-center space-x-2.5">
                    <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-sm shadow-md">
                        <i id="auth-modal-icon" class="fa-solid fa-right-to-bracket"></i>
                    </div>
                    <div>
                        <h3 id="auth-modal-title" class="text-base font-bold text-white">Masuk Akun</h3>
                        <p id="auth-modal-subtitle" class="text-xs text-zinc-400">Simpan riwayat percakapan kamu</p>
                    </div>
                </div>
                <button onclick="closeAuthModal()" class="text-zinc-500 hover:text-zinc-300 p-1.5 rounded-lg bg-zinc-950 border border-zinc-800">
                    <i class="fa-solid fa-xmark text-sm"></i>
                </button>
            </div>

            <!-- Auth Form -->
            <form id="auth-form" onsubmit="handleAuthSubmit(event)" class="space-y-4">
                <div id="field-fullname" class="hidden">
                    <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-1.5">Nama Lengkap</label>
                    <input type="text" id="auth-fullname" placeholder="Masukkan nama Anda..." class="w-full bg-zinc-950 border border-zinc-800 text-white text-sm rounded-xl px-4 py-2.5 focus:outline-none focus:border-blue-500">
                </div>

                <div>
                    <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-1.5">Username</label>
                    <input type="text" id="auth-username" required placeholder="Masukkan username..." class="w-full bg-zinc-950 border border-zinc-800 text-white text-sm rounded-xl px-4 py-2.5 focus:outline-none focus:border-blue-500">
                </div>

                <div>
                    <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-1.5">Password</label>
                    <div class="relative flex items-center">
                        <input type="password" id="auth-password" required placeholder="Masukkan password..." class="w-full bg-zinc-950 border border-zinc-800 text-white text-sm rounded-xl pl-4 pr-10 py-2.5 focus:outline-none focus:border-blue-500">
                        <button type="button" onclick="togglePasswordVisibility('auth-password', 'eye-icon-auth')" class="absolute right-3 text-zinc-400 hover:text-white transition p-1">
                            <i id="eye-icon-auth" class="fa-solid fa-eye"></i>
                        </button>
                    </div>
                </div>

                <div class="flex items-center justify-between pt-1">
                    <label class="flex items-center space-x-2 text-xs text-zinc-400 cursor-pointer hover:text-zinc-200">
                        <input type="checkbox" id="auth-remember-me" checked class="rounded bg-zinc-950 border-zinc-800 text-blue-600 focus:ring-0">
                        <span>Ingat Saya (Remember Me)</span>
                    </label>
                </div>

                <div id="auth-error" class="hidden p-3 bg-red-950/50 border border-red-800/50 text-red-400 text-xs rounded-xl"></div>

                <button type="submit" id="auth-submit-btn" class="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium text-sm py-2.5 rounded-xl transition shadow-md flex items-center justify-center gap-2">
                    <span id="auth-btn-text">Masuk Sekarang</span>
                    <i class="fa-solid fa-arrow-right text-xs"></i>
                </button>
            </form>

            <div class="text-center pt-2 border-t border-zinc-800">
                <p id="auth-switch-text" class="text-xs text-zinc-400">
                    Belum punya akun? <a href="#" onclick="toggleAuthType(event)" class="text-blue-400 hover:underline font-medium">Daftar Akun Baru</a>
                </p>
            </div>
        </div>
    </div>

    
    <!-- Modal QRIS Payment Mayar -->
    <div id="qris-payment-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-4xl p-6 shadow-2xl space-y-4">
            <!-- Modal Header -->
            <div class="flex items-center justify-between pb-3 border-b border-zinc-800">
                <div class="flex items-center space-x-2.5">
                    <div class="w-8 h-8 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center text-sm shadow-md">
                        <i class="fa-solid fa-qrcode"></i>
                    </div>
                    <div>
                        <h3 class="text-sm font-bold text-white">Pembelian Paket Kredit QRIS</h3>
                        <p class="text-[11px] text-zinc-400">Pilih paket kredit harian untuk meningkatkan kuota obrolan AI kamu</p>
                    </div>
                </div>
                <button onclick="closeQRISModal()" class="text-zinc-400 hover:text-white p-1 transition">
                    <i class="fa-solid fa-xmark text-sm"></i>
                </button>
            </div>

            <!-- Step 1: Selection Container -->
            <div id="qris-select-step" class="space-y-3">
                <!-- Dynamic Package Items -->
            </div>

            <!-- Step 2: QRIS Display & Polling Container -->
            <div id="qris-display-step" class="hidden space-y-4 text-center">
                <div class="bg-zinc-950 border border-zinc-800 p-5 rounded-2xl space-y-4 shadow-inner">
                    <div class="space-y-1">
                        <span class="px-2.5 py-0.5 text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full">QRIS DYNAMIC MAYAR.ID</span>
                        <h4 id="qris-tier-title" class="text-sm font-bold text-white pt-1">Bronze Tier - Rp 15.000</h4>
                    </div>

                    <div class="flex justify-center">
                        <div class="p-2.5 bg-white rounded-2xl shadow-xl border border-zinc-700">
                            <img id="qris-img" src="" alt="Dynamic QRIS Code" class="w-52 h-52 object-contain">
                        </div>
                    </div>

                    <div class="bg-zinc-900 border border-zinc-800/80 p-3 rounded-xl space-y-1.5 text-left text-xs">
                        <div class="font-semibold text-zinc-300 flex items-center gap-1.5">
                            <i class="fa-solid fa-wallet text-emerald-400"></i>
                            <span>Metode Pembayaran Resmi:</span>
                        </div>
                        <p class="text-[11px] text-zinc-400 leading-relaxed">Dapat di-scan menggunakan seluruh m-Banking (BCA, Mandiri, BRI, BNI, CIMB) & E-Wallet (GoPay, OVO, Dana, ShopeePay, LinkAja).</p>
                    </div>
                </div>

                <div class="flex items-center justify-center gap-2 text-xs text-amber-400 font-semibold bg-amber-500/10 border border-amber-500/20 py-2.5 px-4 rounded-xl">
                    <i class="fa-solid fa-spinner animate-spin text-sm"></i>
                    <span>Menunggu Pembayaran (Sistem me-refresh otomatis)...</span>
                </div>
            </div>
        </div>
    </div>

    <!-- Modal Profile & ChatGPT-style Billing Dashboard -->
    <div id="profile-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-3xl w-full max-w-2xl p-6 shadow-2xl space-y-5">
            <!-- Modal Header -->
            <div class="flex items-center justify-between pb-4 border-b border-zinc-800">
                <div class="flex items-center space-x-3">
                    <div id="prof-modal-avatar" class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center text-base font-bold shadow-lg shadow-blue-500/20">U</div>
                    <div>
                        <div class="flex items-center gap-2">
                            <h3 id="prof-modal-name" class="text-base font-bold text-white">Nama Pengguna</h3>
                            <span id="prof-modal-role" class="px-2 py-0.5 text-[10px] font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full uppercase">Free Plan</span>
                        </div>
                        <p id="prof-modal-username" class="text-xs text-zinc-400">@username</p>
                    </div>
                </div>
                <button onclick="closeProfileModal()" class="text-zinc-400 hover:text-white p-1.5 rounded-xl bg-zinc-950 border border-zinc-800 transition">
                    <i class="fa-solid fa-xmark text-sm"></i>
                </button>
            </div>

            <!-- Dashboard Navigation Tabs: Billing vs History -->
            <div class="flex items-center bg-zinc-950 border border-zinc-800 p-1 rounded-2xl">
                <button id="prof-tab-btn-billing" onclick="switchProfTab('billing')" class="flex-1 py-2 text-xs font-semibold rounded-xl bg-blue-600 text-white transition flex items-center justify-center gap-1.5 shadow-sm">
                    <i class="fa-solid fa-sparkles text-amber-300 text-xs"></i>
                    <span>Langganan & Kuota</span>
                </button>
                <button id="prof-tab-btn-history" onclick="switchProfTab('history')" class="flex-1 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5">
                    <i class="fa-solid fa-receipt text-xs"></i>
                    <span>Riwayat Transaksi</span>
                </button>
            </div>

            <!-- Tab 1: Billing & Limits -->
            <div id="prof-tab-content-billing" class="space-y-4">
                <!-- Usage Meter Card -->
                <div class="p-4 bg-zinc-950 border border-zinc-800/90 rounded-2xl space-y-3">
                    <div class="flex items-center justify-between">
                        <div>
                            <span class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Penggunaan Kredit Hari Ini</span>
                            <div class="flex items-baseline gap-2 mt-0.5">
                                <span id="prof-modal-usage-count" class="text-lg font-bold text-white font-mono">1 / 5</span>
                                <span class="text-xs text-zinc-400">chat terpakai</span>
                            </div>
                        </div>
                        <div class="text-right">
                            <span class="text-[10px] text-zinc-500 uppercase font-semibold">Sisa Kredit</span>
                            <div id="prof-modal-remaining-count" class="text-sm font-bold text-emerald-400 font-mono">4 Chat</div>
                        </div>
                    </div>

                    <!-- Progress Bar -->
                    <div class="w-full bg-zinc-900 h-2.5 rounded-full overflow-hidden border border-zinc-800/80">
                        <div id="prof-modal-progress-bar" class="bg-gradient-to-r from-blue-500 to-indigo-500 h-full transition-all duration-300" style="width: 20%;"></div>
                    </div>

                    <div class="flex items-center justify-between text-[11px] text-zinc-400 pt-1">
                        <span class="flex items-center gap-1"><i class="fa-regular fa-clock text-blue-400"></i> Auto Reset Jam 00:00 WIB</span>
                        <span id="prof-modal-limit-tag" class="font-mono text-zinc-300">Batas: 5 Chat/Hari</span>
                    </div>
                </div>

                <!-- CTA Banner Upgrade -->
                <div class="p-4 bg-gradient-to-r from-blue-950/40 via-indigo-950/30 to-zinc-950 border border-blue-500/30 rounded-2xl flex items-center justify-between gap-4 shadow-xl">
                    <div class="space-y-1">
                        <h4 class="text-xs font-bold text-white flex items-center gap-1.5">
                            <i class="fa-solid fa-bolt text-amber-400"></i>
                            <span>Butuh Kuota Obrolan AI Lebih Banyak?</span>
                        </h4>
                        <p class="text-[11px] text-zinc-400 leading-relaxed">Nikmati hingga 200 chat/hari dengan mengaktifkan paket langganan resmi via QRIS instant.</p>
                    </div>
                    <button onclick="openPackageModal(); closeProfileModal();" class="px-4 py-2.5 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-blue-500/20 transition shrink-0 flex items-center gap-1.5 active:scale-95">
                        <i class="fa-solid fa-qrcode"></i>
                        <span>Pilih Paket QRIS</span>
                    </button>
                </div>
            </div>

            <!-- Tab 2: Transaction History -->
            <div id="prof-tab-content-history" class="space-y-3 hidden">
                <div class="flex items-center justify-between px-1">
                    <span class="text-xs font-bold text-zinc-300 flex items-center gap-1.5">
                        <i class="fa-solid fa-receipt text-blue-400"></i>
                        <span>Daftar Transaksi Saya</span>
                    </span>
                    <button onclick="loadUserTransactions()" class="text-xs text-blue-400 hover:text-blue-300 font-medium flex items-center gap-1">
                        <i class="fa-solid fa-rotate-right text-[10px]"></i> Refresh
                    </button>
                </div>

                <div id="prof-tx-history-box" class="max-h-72 overflow-y-auto space-y-2 pr-1">
                    <p class="text-xs text-zinc-500 text-center py-6">Memuat riwayat transaksi...</p>
                </div>
            </div>

            <!-- Modal Footer -->
            <div class="flex justify-end gap-2 pt-2 border-t border-zinc-800">
                <button onclick="closeProfileModal()" class="px-5 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-semibold py-2.5 rounded-xl transition">
                    Tutup
                </button>
            </div>
        </div>
    </div>

    <!-- Modal Document Preview -->
    <div id="preview-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-2 md:p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-5xl h-[92vh] md:h-[90vh] flex flex-col shadow-2xl overflow-hidden">
            <div class="p-3 md:p-4 border-b border-zinc-800 flex flex-wrap justify-between items-center bg-zinc-950 gap-2">
                <div class="flex items-center space-x-2.5 min-w-0">
                    <div id="preview-icon-box" class="w-8 h-8 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 flex items-center justify-center shrink-0">
                        <i id="preview-icon" class="fa-solid fa-file-pdf text-sm"></i>
                    </div>
                    <div class="min-w-0">
                        <h3 id="preview-modal-title" class="text-xs md:text-sm font-semibold text-zinc-100 truncate max-w-[140px] sm:max-w-xs md:max-w-md">Preview Dokumen</h3>
                        <p id="preview-modal-subtitle" class="text-[9px] md:text-[10px] text-zinc-400 truncate">PDF Reader Viewer & PDF Download</p>
                    </div>
                </div>

                <div class="flex items-center bg-red-500/10 border border-red-500/20 px-3 py-1 rounded-xl text-red-400 text-xs font-medium gap-1.5 shrink-0">
                    <i class="fa-solid fa-file-pdf"></i>
                    <span>PDF Viewer</span>
                </div>

                <div class="flex items-center space-x-2 shrink-0">
                    <a id="preview-download-btn" href="#" target="_blank" class="text-[11px] md:text-xs bg-zinc-800 hover:bg-zinc-700 text-zinc-200 px-2.5 md:px-3 py-1.5 rounded-lg border border-zinc-700 transition flex items-center gap-1.5">
                        <i class="fa-solid fa-download text-[10px]"></i>
                        <span class="hidden sm:inline">Buka Tab Baru</span>
                    </a>
                    <button onclick="closePreviewModal()" class="text-zinc-400 hover:text-white w-7 h-7 md:w-8 md:h-8 rounded-lg flex items-center justify-center bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 transition">
                        <i class="fa-solid fa-xmark text-sm"></i>
                    </button>
                </div>
            </div>

            <div class="flex-1 bg-zinc-950 relative overflow-hidden flex flex-col">
                <iframe id="preview-iframe" class="w-full h-full border-0 hidden" src=""></iframe>
                <div id="preview-txt-container" class="w-full h-full p-4 md:p-6 overflow-y-auto hidden">
                    <pre id="preview-txt-content" class="text-[11px] md:text-xs text-zinc-300 font-mono whitespace-pre-wrap leading-relaxed select-text"></pre>
                </div>
                <div id="preview-modal-loading" class="flex flex-col items-center justify-center h-full space-y-3 py-12">
                    <i class="fa-solid fa-spinner animate-spin text-2xl md:text-3xl text-blue-500"></i>
                    <span id="preview-loading-text" class="text-xs text-zinc-400">Memuat berkas...</span>
                </div>
            </div>
        </div>
    </div>

    <script>
        function togglePasswordVisibility(inputId, iconId) {
            const input = document.getElementById(inputId);
            const icon = document.getElementById(iconId);
            if (!input || !icon) return;

            if (input.type === 'password') {
                input.type = 'text';
                icon.className = 'fa-solid fa-eye-slash';
            } else {
                input.type = 'password';
                icon.className = 'fa-solid fa-eye';
            }
        }

        let currentUser = null;
        let activeSessionID = "";
        let authMode = "login"; // 'login' or 'register'

        // Check Auth Status on Load
        document.addEventListener("DOMContentLoaded", () => {
            checkAuth();
        });

        async function checkAuth() {
            try {
                const res = await fetch("/api/auth/me");
                const data = await res.json();
                if (data.authenticated && data.user) {
                    currentUser = data.user;
                    renderUserNav();
                    loadUserSessions();
                } else {
                    currentUser = null;
                    renderUserNav();
                }
            } catch (err) {
                currentUser = null;
                renderUserNav();
            }
        }

        function renderUserNav() {
            const container = document.getElementById("auth-container");
            const historyNotice = document.getElementById("history-login-notice");
            const historyList = document.getElementById("history-sessions-list");
            const quizAuthBanner = document.getElementById("quiz-auth-banner");
            const quizMainContent = document.getElementById("quiz-main-content");
            const sbUserFooter = document.getElementById("sidebar-user-footer");

            const chatLock = document.getElementById("chat-guest-lock");
            const chatWrapper = document.getElementById("chat-input-wrapper");

            if (currentUser) {
                if (chatLock) chatLock.classList.add("hidden");
                if (chatWrapper) chatWrapper.classList.remove("hidden");
                const name = currentUser.full_name || currentUser.username;
                const initial = name ? name[0].toUpperCase() : 'U';
                let creditText = '';
                let chatCreditText = '';
                let planBadgeStr = 'Free Plan';

                if (currentUser.role === 'admin') {
                    creditText = 'Unlimited Admin';
                    chatCreditText = '⚡ Kuota Chat: Unlimited (Admin)';
                    planBadgeStr = 'Admin';
                } else if (currentUser.daily_limit === 50) {
                    planBadgeStr = 'Bronze';
                } else if (currentUser.daily_limit === 100) {
                    planBadgeStr = 'Silver';
                } else if (currentUser.daily_limit === 200) {
                    planBadgeStr = 'Gold';
                }

                if (currentUser.role !== 'admin') {
                    const remaining = typeof currentUser.remaining_today !== 'undefined' ? currentUser.remaining_today : currentUser.daily_limit;
                    creditText = 'Sisa Kredit: ' + remaining + '/' + currentUser.daily_limit + ' Chat';
                    chatCreditText = '⚡ Sisa Kuota Hari Ini: ' + remaining + ' / ' + currentUser.daily_limit + ' Chat';
                }

                const chatIndicator = document.getElementById("chat-credit-indicator");
                if (chatIndicator) chatIndicator.innerText = chatCreditText;

                // Navbar User Info
                container.innerHTML = '<div class="flex items-center space-x-2">' +
                    '<span id="credit-badge" onclick="openProfileModal(); switchProfTab(\'billing\');" class="px-2.5 py-1 text-[11px] font-medium bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border border-blue-500/20 rounded-lg flex items-center gap-1.5 shadow-sm cursor-pointer transition">' +
                        '<i class="fa-solid fa-bolt text-[10px]"></i>' +
                        '<span>' + escapeHtml(creditText) + '</span>' +
                    '</span>' +
                    '<div class="flex items-center space-x-2 bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 px-3 py-1 rounded-xl cursor-pointer transition" onclick="openProfileModal()">' +
                        '<div class="w-6 h-6 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center text-xs font-bold shadow-sm">' + initial + '</div>' +
                        '<span class="text-xs font-semibold text-zinc-200 max-w-[100px] truncate">' + escapeHtml(name) + '</span>' +
                    '</div>' +
                    '<button onclick="logoutUser()" title="Keluar" class="text-zinc-400 hover:text-rose-400 p-1.5 transition">' +
                        '<i class="fa-solid fa-right-from-bracket text-xs"></i>' +
                    '</button>' +
                '</div>';

                // Sidebar User Widget Sync
                if (sbUserFooter) {
                    sbUserFooter.classList.remove("hidden");
                    const sbAvatar = document.getElementById("sb-user-avatar");
                    const sbName = document.getElementById("sb-user-name");
                    const sbPlanBadge = document.getElementById("sb-user-plan-badge");
                    const sbUsageStr = document.getElementById("sb-user-usage-str");

                    if (sbAvatar) sbAvatar.innerText = initial;
                    if (sbName) sbName.innerText = name;
                    if (sbPlanBadge) sbPlanBadge.innerText = planBadgeStr;
                    if (sbUsageStr) {
                        const used = currentUser.used_today !== undefined ? currentUser.used_today : 0;
                        sbUsageStr.innerText = currentUser.role === 'admin' ? 'Unlimited' : (used + '/' + currentUser.daily_limit);
                    }
                }

                if (historyNotice) historyNotice.classList.add("hidden");
                if (historyList) historyList.classList.remove("hidden");
                if (quizAuthBanner) quizAuthBanner.classList.add("hidden");
                if (quizMainContent) quizMainContent.classList.remove("hidden");
            } else {
                container.innerHTML = '<button onclick="openAuthModal(\'login\')" class="text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white px-3 py-1.5 rounded-lg transition shadow-sm flex items-center gap-1.5">' +
                        '<i class="fa-solid fa-right-to-bracket text-[10px]"></i>' +
                        '<span>Masuk</span>' +
                    '</button>' +
                    '<button onclick="openAuthModal(\'register\')" class="text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 px-3 py-1.5 rounded-lg border border-zinc-700 transition flex items-center gap-1.5 shadow-sm">' +
                        '<i class="fa-solid fa-user-plus text-[10px]"></i>' +
                        '<span>Daftar</span>' +
                    '</button>';
                if (sbUserFooter) sbUserFooter.classList.add("hidden");
                if (historyNotice) historyNotice.classList.remove("hidden");
                if (historyList) {
                    historyList.classList.add("hidden");
                    historyList.innerHTML = "";
                }
                if (quizAuthBanner) quizAuthBanner.classList.remove("hidden");
                if (quizMainContent) quizMainContent.classList.add("hidden");
                if (chatLock) chatLock.classList.remove("hidden");
                if (chatWrapper) chatWrapper.classList.add("hidden");
            }
            loadLeaderboard();
        }

        async function logoutUser() {
            await fetch("/api/auth/logout", { method: "POST" });
            currentUser = null;
            activeSessionID = "";
            renderUserNav();
            // Reset chat box
            document.getElementById("chat-box").innerHTML = '<div class="flex items-start gap-3">' +
                    '<div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xs shrink-0 shadow-md">' +
                        '<i class="fa-solid fa-sparkles"></i>' +
                    '</div>' +
                    '<div class="aui-msg-assistant p-4 text-sm text-zinc-200 leading-relaxed shadow-sm">' +
                        'Halo! Saya Asisten AI Kurikulum Koding & AI SD-SMA. Silakan tanyakan materi kurikulum!' +
                    '</div>' +
                '</div>';
        }

        
        let currentTxID = null;
        let pollTimer = null;

        function closeQRISModal() {
            document.getElementById('qris-payment-modal').classList.add('hidden');
            if (pollTimer) {
                clearInterval(pollTimer);
                pollTimer = null;
            }
            currentTxID = null;
        }

        async function openPackageModal() {
            if (!currentUser) return openAuthModal('login');
            const stepSelect = document.getElementById('qris-select-step');
            const stepDisplay = document.getElementById('qris-display-step');

            stepSelect.classList.remove('hidden');
            stepDisplay.classList.add('hidden');
            stepSelect.innerHTML = '<p class="text-xs text-zinc-500 py-4 text-center">Memuat daftar paket kredit...</p>';
            document.getElementById('qris-payment-modal').classList.remove('hidden');

            try {
                const res = await fetch('/api/packages');
                const data = await res.json();

                if (!data.is_gateway_configured) {
                    stepSelect.innerHTML = '<div class="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl space-y-2 text-center">' +
                        '<div class="text-amber-400 font-bold text-xs flex items-center justify-center gap-1.5"><i class="fa-solid fa-triangle-exclamation"></i> Gateway Pembayaran Belum Siap</div>' +
                        '<p class="text-xs text-zinc-300">Admin belum mengaktifkan API Key Mayar.id. Fitur topup / pembelian paket QRIS sementara belum tersedia.</p>' +
                    '</div>';
                    return;
                }

                if (!data.packages || data.packages.length === 0) {
                    stepSelect.innerHTML = '<p class="text-xs text-zinc-500 py-4 text-center">Belum ada paket kredit yang tersedia.</p>';
                    return;
                }

                let currentLimit = (currentUser && currentUser.daily_limit) ? currentUser.daily_limit : 5;
                let html = '<div class="space-y-4">' +
                    '<div class="text-center space-y-1">' +
                        '<h4 class="text-base font-bold text-white">Pilih Paket Kredit Obrolan AI</h4>' +
                        '<p class="text-xs text-zinc-400">Tingkatkan limit harian kamu dengan berlangganan paket resmi Assistant-UI Base</p>' +
                    '</div>' +
                    '<div class="grid grid-cols-1 md:grid-cols-3 gap-3.5 pt-2">';

                data.packages.forEach(p => {
                    let isFeatured = p.daily_limit >= 100 && p.daily_limit < 500;
                    let isCurrent = p.daily_limit === currentLimit;
                    let isUpgrade = p.daily_limit > currentLimit;
                    
                    let cardClass = isCurrent ? 'aui-pricing-card border-blue-500/60 bg-blue-950/20 shadow-lg ring-1 ring-blue-500/30' : (isFeatured ? 'aui-pricing-card featured' : 'aui-pricing-card');

                    let badgeHtml = '';
                    if (isCurrent) {
                        badgeHtml = '<span class="px-2 py-0.5 text-[9px] font-bold bg-blue-500 text-white rounded-full uppercase tracking-wider">Paket Aktif</span>';
                    } else if (isUpgrade) {
                        badgeHtml = '<span class="px-2 py-0.5 text-[9px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full uppercase tracking-wider">Upgrade ⬆</span>';
                    } else {
                        badgeHtml = '<span class="px-2 py-0.5 text-[9px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-full uppercase tracking-wider">Downgrade ⬇</span>';
                    }

                    html += '<div class="' + cardClass + ' p-4 flex flex-col justify-between space-y-4">' +
                        '<div class="space-y-2.5">' +
                            '<div class="flex items-center justify-between">' +
                                '<span class="text-xs font-bold text-white uppercase tracking-wider">' + escapeHtml(p.name) + '</span>' +
                                badgeHtml +
                            '</div>' +
                            '<div class="flex items-baseline gap-1">' +
                                '<span class="text-xl font-bold text-white font-mono">Rp ' + (p.price/1000).toFixed(0) + 'k</span>' +
                                '<span class="text-[10px] text-zinc-400">/ bulan</span>' +
                            '</div>' +
                            '<p class="text-[11px] text-zinc-400 leading-relaxed border-t border-zinc-800/80 pt-2">' + escapeHtml(p.description || '-') + '</p>' +
                            '<ul class="space-y-1.5 text-[11px] text-zinc-300 pt-1">' +
                                '<li class="flex items-center gap-2"><i class="fa-solid fa-check text-blue-400 text-[10px]"></i> <span><strong>' + p.daily_limit + ' Chat</strong> / Hari</span></li>' +
                                '<li class="flex items-center gap-2"><i class="fa-solid fa-check text-blue-400 text-[10px]"></i> <span>Auto Reset Jam 00:00</span></li>' +
                                '<li class="flex items-center gap-2"><i class="fa-solid fa-check text-blue-400 text-[10px]"></i> <span>Direct Vision AI PDF</span></li>' +
                            '</ul>' +
                        '</div>' +
                        '<button onclick="selectDynamicPackage(' + p.id + ')" class="w-full bg-blue-600 hover:bg-blue-500 active:scale-95 text-white text-xs font-semibold py-2.5 rounded-xl transition shadow-md flex items-center justify-center gap-1.5">' +
                            '<i class="fa-solid fa-qrcode"></i> Bayar QRIS' +
                        '</button>' +
                    '</div>';
                });
                html += '</div></div>';
                stepSelect.innerHTML = html;

            } catch(e) {
                stepSelect.innerHTML = '<p class="text-xs text-red-400 py-4 text-center">Gagal memuat daftar paket.</p>';
            }
        }

        async function selectDynamicPackage(packageID) {
            try {
                const res = await fetch('/api/payment/create-qris', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({package_id: packageID})
                });
                const data = await res.json();
                if (res.ok && data.transaction) {
                    currentTxID = data.transaction.id;
                    document.getElementById('qris-tier-title').innerText = 'Paket ' + data.transaction.tier_name + ' Tier - Rp ' + data.transaction.amount.toLocaleString('id-ID');
                    document.getElementById('qris-img').src = data.transaction.qr_url;

                    document.getElementById('qris-select-step').classList.add('hidden');
                    document.getElementById('qris-display-step').classList.remove('hidden');

                    if (pollTimer) clearInterval(pollTimer);
                    pollTimer = setInterval(checkPaymentStatus, 3000);
                } else {
                    alert('Gagal generate QRIS: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server pembayaran.');
            }
        }

        async function checkPaymentStatus() {
            if (!currentTxID) return;
            try {
                const res = await fetch('/api/payment/status?tx_id=' + currentTxID);
                const data = await res.json();
                if (res.ok && data.transaction && data.transaction.status === 'paid') {
                    clearInterval(pollTimer);
                    alert('Pembayaran Sukses! Paket kredit harian kamu telah aktif!');
                    closeQRISModal();
                    checkAuth();
                }
            } catch(e) {}
        }

        
        async function cancelTxHistory(txID) {
            if (!confirm('Apakah Anda yakin ingin membatalkan transaksi pending ini?')) return;
            try {
                const res = await fetch('/api/payment/cancel', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({tx_id: txID})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    loadUserTransactions();
                } else {
                    alert('Gagal membatalkan transaksi: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server.');
            }
        }

        async function deleteTxHistory(txID) {
            if (!confirm('Hapus transaksi ini dari riwayat?')) return;
            try {
                const res = await fetch('/api/payment/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({tx_id: txID})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    loadUserTransactions();
                } else {
                    alert('Gagal menghapus riwayat: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server.');
            }
        }

        function switchProfTab(tab) {
            const btnBilling = document.getElementById("prof-tab-btn-billing");
            const btnHistory = document.getElementById("prof-tab-btn-history");
            const contentBilling = document.getElementById("prof-tab-content-billing");
            const contentHistory = document.getElementById("prof-tab-content-history");

            if (tab === "billing") {
                if (btnBilling) btnBilling.className = "flex-1 py-2 text-xs font-semibold rounded-xl bg-blue-600 text-white transition flex items-center justify-center gap-1.5 shadow-sm";
                if (btnHistory) btnHistory.className = "flex-1 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5";
                if (contentBilling) contentBilling.classList.remove("hidden");
                if (contentHistory) contentHistory.classList.add("hidden");
            } else {
                if (btnHistory) btnHistory.className = "flex-1 py-2 text-xs font-semibold rounded-xl bg-blue-600 text-white transition flex items-center justify-center gap-1.5 shadow-sm";
                if (btnBilling) btnBilling.className = "flex-1 py-2 text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5";
                if (contentHistory) contentHistory.classList.remove("hidden");
                if (contentBilling) contentBilling.classList.add("hidden");
                loadUserTransactions();
            }
        }

        async function loadUserTransactions() {
            const container = document.getElementById("prof-tx-history-box");
            if (!container) return;
            container.innerHTML = '<div class="flex items-center justify-center py-6 text-zinc-500 text-xs gap-2"><i class="fa-solid fa-spinner animate-spin"></i> Memuat riwayat transaksi...</div>';
            try {
                const res = await fetch("/api/payment/transactions");
                const data = await res.json();
                if (res.ok && data.transactions && data.transactions.length > 0) {
                    let html = '<div class="space-y-2">';
                    data.transactions.forEach(t => {
                        let statusBadge = "";
                        let actionBtns = "";

                        if (t.status === "paid") {
                            statusBadge = '<span class="px-2 py-0.5 text-[9px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-md flex items-center gap-1"><i class="fa-solid fa-circle-check"></i> LUNAS</span>';
                        } else if (t.status === "pending") {
                            statusBadge = '<span class="px-2 py-0.5 text-[9px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-md flex items-center gap-1"><i class="fa-solid fa-clock"></i> PENDING</span>';
                            actionBtns = '<div class="flex items-center gap-2 pt-2 border-t border-zinc-900">' +
                                '<button onclick="resumePayment(\'' + t.id + '\', \'' + escapeHtml(t.tier_name) + '\', ' + t.amount + ', \'' + escapeHtml(t.qr_url) + '\')" class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-[11px] font-semibold rounded-lg transition flex items-center gap-1.5 shadow-sm active:scale-95">' +
                                    '<i class="fa-solid fa-qrcode"></i> Bayar QRIS' +
                                '</button>' +
                                '<button onclick="cancelTxHistory(\'' + t.id + '\')" class="px-2.5 py-1.5 bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800 text-[11px] font-medium rounded-lg transition flex items-center gap-1">' +
                                    '<i class="fa-solid fa-ban text-rose-400 text-[10px]"></i> Batalkan' +
                                '</button>' +
                            '</div>';
                        } else {
                            statusBadge = '<span class="px-2 py-0.5 text-[9px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20 rounded-md flex items-center gap-1"><i class="fa-solid fa-circle-xmark"></i> EXPIRED</span>';
                            actionBtns = '<div class="pt-2 border-t border-zinc-900">' +
                                '<button onclick="deleteTxHistory(\'' + t.id + '\')" class="px-2.5 py-1 bg-zinc-900 hover:bg-rose-950/50 text-rose-400 border border-zinc-800 text-[10px] rounded-lg transition flex items-center gap-1">' +
                                    '<i class="fa-solid fa-trash-can"></i> Hapus' +
                                '</button>' +
                            '</div>';
                        }

                        let formattedDate = new Date(t.created_at).toLocaleDateString("id-ID", {day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit"});

                        html += '<div class="p-3 bg-zinc-950 border border-zinc-800/80 rounded-2xl space-y-2 text-xs transition hover:border-zinc-700 shadow-sm">' +
                            '<div class="flex items-center justify-between">' +
                                '<div class="flex items-center gap-2">' +
                                    '<span class="font-bold text-zinc-100 text-xs">' + escapeHtml(t.tier_name) + ' Tier</span>' +
                                    statusBadge +
                                '</div>' +
                                '<span class="font-mono font-bold text-emerald-400 text-xs">Rp ' + t.amount.toLocaleString("id-ID") + '</span>' +
                            '</div>' +
                            '<div class="flex items-center justify-between text-[10px] text-zinc-500 font-mono">' +
                                '<span>ID: ' + t.id.substring(0, 14) + '...</span>' +
                                '<span>' + formattedDate + '</span>' +
                            '</div>' +
                            (actionBtns ? actionBtns : '') +
                        '</div>';
                    });
                    html += '</div>';
                    container.innerHTML = html;
                } else {
                    container.innerHTML = '<div class="p-8 text-center space-y-2 border border-dashed border-zinc-800 rounded-2xl">' +
                        '<i class="fa-solid fa-receipt text-3xl text-zinc-600"></i>' +
                        '<p class="text-xs text-zinc-300 font-bold">Belum Ada Riwayat Transaksi</p>' +
                        '<p class="text-[11px] text-zinc-500">Pilih paket langganan untuk meningkatkan kuota obrolan harian Anda.</p>' +
                    '</div>';
                }
            } catch(e) {
                container.innerHTML = '<p class="text-xs text-rose-400 text-center py-4">Gagal memuat riwayat transaksi.</p>';
            }
        }

        function resumePayment(txID, tierName, amount, qrURL) {
            closeProfileModal();
            currentTxID = txID;
            document.getElementById('qris-tier-title').innerText = 'Paket ' + tierName + ' Tier - Rp ' + amount.toLocaleString('id-ID');
            document.getElementById('qris-img').src = qrURL;

            document.getElementById('qris-select-step').classList.add('hidden');
            document.getElementById('qris-display-step').classList.remove('hidden');
            document.getElementById('qris-payment-modal').classList.remove('hidden');

            if (pollTimer) clearInterval(pollTimer);
            pollTimer = setInterval(checkPaymentStatus, 3000);
        }

        function openProfileModal() {
            if (!currentUser) return openAuthModal('login');
            const name = currentUser.full_name || currentUser.username;
            const initial = name ? name[0].toUpperCase() : 'U';

            const avatarEl = document.getElementById("prof-modal-avatar");
            const nameEl = document.getElementById("prof-modal-name");
            const usernameEl = document.getElementById("prof-modal-username");
            const roleEl = document.getElementById("prof-modal-role");
            const usageCountEl = document.getElementById("prof-modal-usage-count");
            const remainingCountEl = document.getElementById("prof-modal-remaining-count");
            const progressBarEl = document.getElementById("prof-modal-progress-bar");
            const limitTagEl = document.getElementById("prof-modal-limit-tag");

            if (avatarEl) avatarEl.innerText = initial;
            if (nameEl) nameEl.innerText = name;
            if (usernameEl) usernameEl.innerText = "@" + currentUser.username;

            let roleText = "Free Plan";
            if (currentUser.role === "admin") {
                roleText = "Unlimited Admin";
            } else if (currentUser.daily_limit === 50) {
                roleText = "Bronze Tier";
            } else if (currentUser.daily_limit === 100) {
                roleText = "Silver Tier";
            } else if (currentUser.daily_limit === 200) {
                roleText = "Gold Tier";
            } else if (currentUser.daily_limit > 5) {
                roleText = "Custom Tier";
            }
            if (roleEl) roleEl.innerText = roleText;

            const limit = currentUser.daily_limit || 5;
            const used = currentUser.used_today !== undefined ? currentUser.used_today : 0;
            const remaining = currentUser.remaining_today !== undefined ? currentUser.remaining_today : (limit - used);

            if (currentUser.role === "admin") {
                if (usageCountEl) usageCountEl.innerText = "Unlimited";
                if (remainingCountEl) remainingCountEl.innerText = "Unlimited";
                if (progressBarEl) progressBarEl.style.width = "100%";
                if (limitTagEl) limitTagEl.innerText = "Batas: Unlimited";
            } else {
                if (usageCountEl) usageCountEl.innerText = used + " / " + limit;
                if (remainingCountEl) remainingCountEl.innerText = remaining + " Chat";
                const percentage = Math.min(100, Math.max(0, (used / limit) * 100));
                if (progressBarEl) progressBarEl.style.width = percentage + "%";
                if (limitTagEl) limitTagEl.innerText = "Batas: " + limit + " Chat/Hari";
            }

            switchProfTab('billing');
            document.getElementById("profile-modal").classList.remove("hidden");
        }

        function closeProfileModal() {
            document.getElementById("profile-modal").classList.add("hidden");
        }

        function openAuthModal(mode) {
            authMode = mode;
            const modal = document.getElementById("auth-modal");
            const title = document.getElementById("auth-modal-title");
            const subtitle = document.getElementById("auth-modal-subtitle");
            const icon = document.getElementById("auth-modal-icon");
            const fieldFullname = document.getElementById("field-fullname");
            const btnText = document.getElementById("auth-btn-text");
            const switchText = document.getElementById("auth-switch-text");
            const errDiv = document.getElementById("auth-error");

            errDiv.classList.add("hidden");

            if (mode === "register") {
                title.innerText = "Daftar Akun Baru";
                subtitle.innerText = "Buat akun untuk menyimpan riwayat chat";
                icon.className = "fa-solid fa-user-plus";
                fieldFullname.classList.remove("hidden");
                btnText.innerText = "Daftar Akun Baru";
                switchText.innerHTML = 'Sudah punya akun? <a href="#" onclick="toggleAuthType(event)" class="text-blue-400 hover:underline font-medium">Masuk Sekarang</a>';
            } else {
                title.innerText = "Masuk Akun";
                subtitle.innerText = "Simpan dan lanjutkan riwayat percakapan kamu";
                icon.className = "fa-solid fa-right-to-bracket";
                fieldFullname.classList.add("hidden");
                btnText.innerText = "Masuk Sekarang";
                switchText.innerHTML = 'Belum punya akun? <a href="#" onclick="toggleAuthType(event)" class="text-blue-400 hover:underline font-medium">Daftar Akun Baru</a>';
            }
            modal.classList.remove("hidden");
        }

        function closeAuthModal() {
            document.getElementById("auth-modal").classList.add("hidden");
        }

        function toggleAuthType(e) {
            e.preventDefault();
            openAuthModal(authMode === "login" ? "register" : "login");
        }

        async function handleAuthSubmit(e) {
            e.preventDefault();
            const username = document.getElementById("auth-username").value.trim();
            const password = document.getElementById("auth-password").value.trim();
            const fullName = document.getElementById("auth-fullname").value.trim();
            const rememberMe = document.getElementById("auth-remember-me") ? document.getElementById("auth-remember-me").checked : true;
            const errDiv = document.getElementById("auth-error");

            errDiv.classList.add("hidden");
            const endpoint = authMode === "register" ? "/api/auth/register" : "/api/auth/login";
            const payload = authMode === "register" 
                ? { username, password, full_name: fullName } 
                : { username, password, remember_me: rememberMe };

            try {
                const res = await fetch(endpoint, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload)
                });
                const data = await res.json();

                if (res.ok && data.success) {
                    currentUser = data.user;
                    closeAuthModal();
                    if (currentUser && currentUser.role === 'admin') {
                        window.location.href = '/admin';
                    } else {
                        renderUserNav();
                        loadUserSessions();
                    }
                } else {
                    errDiv.classList.remove("hidden");
                    errDiv.innerText = data.error || "Terjadi kesalahan saat otentikasi.";
                }
            } catch (err) {
                errDiv.classList.remove("hidden");
                errDiv.innerText = "Gagal terhubung ke server.";
            }
        }

        // Sidebar Tabs Switcher
        function switchSidebarTab(tab) {
            const btnDocs = document.getElementById("tab-btn-docs");
            const btnHistory = document.getElementById("tab-btn-history");
            const contentDocs = document.getElementById("tab-content-docs");
            const contentHistory = document.getElementById("tab-content-history");

            if (tab === "docs") {
                btnDocs.className = "flex-1 py-1.5 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center justify-center gap-1.5";
                btnHistory.className = "flex-1 py-1.5 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5";
                contentDocs.classList.remove("hidden");
                contentHistory.classList.add("hidden");
            } else {
                btnHistory.className = "flex-1 py-1.5 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center justify-center gap-1.5";
                btnDocs.className = "flex-1 py-1.5 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center justify-center gap-1.5";
                contentHistory.classList.remove("hidden");
                contentDocs.classList.add("hidden");

                if (currentUser) {
                    loadUserSessions();
                }
            }
        }

        // Load & Manage Chat Sessions
        async function loadUserSessions() {
            if (!currentUser) return;
            const container = document.getElementById("history-sessions-list");
            try {
                const res = await fetch("/api/sessions");
                const data = await res.json();
                if (data.sessions) {
                    renderSessionsList(data.sessions);
                }
            } catch (err) {
                console.error("Failed to load sessions:", err);
            }
        }

        function renderSessionsList(sessions) {
            const container = document.getElementById("history-sessions-list");
            const badge = document.getElementById("session-count-badge");
            if (badge) badge.innerText = sessions.length + " Chat";

            container.innerHTML = "";

            if (sessions.length === 0) {
                container.innerHTML = '<div class="p-6 text-center space-y-1.5 text-zinc-500">' +
                    '<i class="fa-regular fa-comments text-2xl"></i>' +
                    '<p class="text-xs font-medium text-zinc-400">Belum ada percakapan</p>' +
                    '<p class="text-[11px]">Klik "Percakapan Baru" di atas untuk mulai obrolan.</p>' +
                '</div>';
                return;
            }

            sessions.forEach(s => {
                const item = document.createElement("div");
                const isActive = s.id === activeSessionID;
                item.className = "p-2.5 rounded-xl border text-xs cursor-pointer transition flex items-center justify-between group " +
                    (isActive ? "bg-blue-600/15 border-blue-500/40 text-blue-300 font-semibold shadow-sm" : "bg-zinc-900/60 hover:bg-zinc-900 border-zinc-800/80 text-zinc-300");

                item.onclick = () => loadSessionMessages(s.id);

                item.innerHTML = '<div class="flex items-center space-x-2.5 min-w-0 flex-1">' +
                        '<i class="fa-regular fa-message text-blue-400 shrink-0 text-xs"></i>' +
                        '<span class="truncate font-medium leading-snug">' + escapeHtml(s.title) + '</span>' +
                    '</div>' +
                    '<button onclick="deleteSession(event, \'' + s.id + '\')" title="Hapus Percakapan" class="opacity-0 group-hover:opacity-100 text-zinc-500 hover:text-rose-400 p-1 transition ml-1 shrink-0">' +
                        '<i class="fa-solid fa-trash-can text-[11px]"></i>' +
                    '</button>';
                container.appendChild(item);
            });
        }

        async function loadSessionMessages(sessionID) {
            activeSessionID = sessionID;
            loadUserSessions();

            try {
                const res = await fetch('/api/session/messages?id=' + encodeURIComponent(sessionID));
                const data = await res.json();
                if (data.messages) {
                    const chatBox = document.getElementById("chat-box");
                    chatBox.innerHTML = "";
                    data.messages.forEach(msg => {
                        appendMessage(msg.role, msg.content);
                    });
                }
            } catch (err) {
                console.error("Failed to load messages:", err);
            }
        }

        async function deleteSession(e, sessionID) {
            e.stopPropagation();
            if (!confirm("Hapus percakapan ini?")) return;

            try {
                const res = await fetch('/api/session/delete?id=' + encodeURIComponent(sessionID), { method: "POST" });
                if (res.ok) {
                    if (activeSessionID === sessionID) {
                        startNewChatSession();
                    } else {
                        loadUserSessions();
                    }
                }
            } catch (err) {
                alert("Gagal menghapus percakapan");
            }
        }

        function startNewChatSession() {
            activeSessionID = "";
            const chatBox = document.getElementById("chat-box");
            chatBox.innerHTML = '<div class="flex items-start gap-3">' +
                    '<div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xs shrink-0 shadow-md">' +
                        '<i class="fa-solid fa-sparkles"></i>' +
                    '</div>' +
                    '<div class="aui-msg-assistant p-4 text-sm text-zinc-200 leading-relaxed shadow-sm">' +
                        'Halo! Obrolan baru dimulai. Silakan tanyakan materi Kurikulum Koding & AI!' +
                    '</div>' +
                '</div>';
            if (currentUser) loadUserSessions();
        }

        function toggleMobileSidebar() {
            const sidebar = document.getElementById('sidebar-docs');
            const backdrop = document.getElementById('mobile-sidebar-backdrop');
            const isHidden = sidebar.classList.contains('-translate-x-full');

            if (isHidden) {
                sidebar.classList.remove('-translate-x-full');
                backdrop.classList.remove('hidden');
            } else {
                sidebar.classList.add('-translate-x-full');
                backdrop.classList.add('hidden');
            }
        }

        let currentRawName = '';
        let currentFilename = '';

        function previewDoc(rawName, filename) {
            currentRawName = rawName;
            currentFilename = filename;

            document.getElementById('preview-modal').classList.remove('hidden');
            renderPreviewContent();
        }

        async function renderPreviewContent() {
            const titleEl = document.getElementById('preview-modal-title');
            const subtitleEl = document.getElementById('preview-modal-subtitle');
            const iconBox = document.getElementById('preview-icon-box');
            const icon = document.getElementById('preview-icon');
            const iframe = document.getElementById('preview-iframe');
            const loadingEl = document.getElementById('preview-modal-loading');
            const loadingText = document.getElementById('preview-loading-text');
            const downloadBtn = document.getElementById('preview-download-btn');

            titleEl.innerText = currentFilename;
            loadingEl.classList.remove('hidden');
            iframe.classList.add('hidden');

            iconBox.className = "w-8 h-8 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 flex items-center justify-center";
            icon.className = "fa-solid fa-file-pdf text-sm";
            subtitleEl.innerText = "Viewer Berkas PDF/DOCX Asli (Visual & Layout)";
            loadingText.innerText = "Memuat berkas PDF asli...";

            const pdfUrl = '/api/doc/preview?file=' + encodeURIComponent(currentRawName);
            downloadBtn.href = pdfUrl;
            downloadBtn.style.display = 'flex';
            iframe.src = pdfUrl;

            iframe.onload = () => {
                loadingEl.classList.add('hidden');
                iframe.classList.remove('hidden');
            };
        }

        function closePreviewModal() {
            const modal = document.getElementById('preview-modal');
            const iframe = document.getElementById('preview-iframe');
            iframe.src = '';
            modal.classList.add('hidden');
        }

        const chatForm = document.getElementById('chat-form');
        const userInput = document.getElementById('user-input');
        const chatBox = document.getElementById('chat-box');

        // Mode Switching
        function switchMode(mode) {
            const chatView = document.getElementById('chat-view');
            const quizView = document.getElementById('quiz-view');
            const btnChat = document.getElementById('btn-mode-chat');
            const btnQuiz = document.getElementById('btn-mode-quiz');

            if (mode === 'chat') {
                chatView.classList.remove('hidden');
                quizView.classList.add('hidden');
                quizView.classList.remove('flex');

                btnChat.className = "px-3 py-1 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center gap-1.5 shadow-sm";
                btnQuiz.className = "px-3 py-1 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center gap-1.5";
            } else {
                chatView.classList.add('hidden');
                quizView.classList.remove('hidden');
                quizView.classList.add('flex');

                btnQuiz.className = "px-3 py-1 rounded-lg text-xs font-medium bg-blue-600 text-white transition flex items-center gap-1.5 shadow-sm";
                btnChat.className = "px-3 py-1 rounded-lg text-xs font-medium text-zinc-400 hover:text-zinc-200 transition flex items-center gap-1.5";

                loadPublicQuizCategories();
                loadLeaderboard();
                if (currentUser && currentQuiz.length === 0) {
                    // Quiz ready for selection
                }
            }
        }

        // Quiz State
        let currentQuiz = [];
        let currentQIndex = 0;
        let userAnswers = {};
        let score = 0;

        async function loadPublicQuizCategories() {
            const select = document.getElementById("quiz-category-select");
            if (!select) return;
            try {
                const res = await fetch("/api/quiz/categories");
                const data = await res.json();
                if (res.ok && data.categories && data.categories.length > 0) {
                    let html = '<option value="0">Pilih Kategori Kuis...</option>';
                    data.categories.forEach(c => {
                        html += '<option value="' + c.id + '">' + escapeHtml(c.name) + ' (' + escapeHtml(c.grade) + ' - ' + c.total_questions + ' Soal)</option>';
                    });
                    select.innerHTML = html;
                } else {
                    select.innerHTML = '<option value="0">Belum ada Kategori Kuis dari Admin</option>';
                }
            } catch(e) {
                select.innerHTML = '<option value="0">Gagal memuat kategori kuis</option>';
            }
        }

        async function generateQuiz() {
            if (!currentUser) {
                openAuthModal('login');
                return;
            }
            const catSelect = document.getElementById('quiz-category-select');
            const catID = parseInt(catSelect ? catSelect.value : "0") || 0;
            if (catID <= 0) {
                return alert('Pilih Kategori Kuis terlebih dahulu!');
            }

            const container = document.getElementById('quiz-container');
            const loading = document.getElementById('quiz-loading');
            const finalResult = document.getElementById('quiz-final-result');

            container.classList.add('hidden');
            finalResult.classList.add('hidden');
            loading.classList.remove('hidden');

            try {
                const res = await fetch('/api/quiz/generate', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({category_id: catID})
                });

                const data = await res.json();
                loading.classList.add('hidden');

                if (res.status === 401) {
                    currentUser = null;
                    renderUserNav();
                    openAuthModal('login');
                    return;
                }

                if (data.questions && data.questions.length > 0) {
                    currentQuiz = data.questions;
                    currentQIndex = 0;
                    userAnswers = {};
                    score = 0;
                    container.classList.remove('hidden');
                    renderQuestion();
                } else {
                    alert('Gagal membuat kuis: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch (err) {
                loading.classList.add('hidden');
                alert('Gagal terhubung ke server kuis.');
            }
        }

        function renderQuestion() {
            const q = currentQuiz[currentQIndex];
            document.getElementById('quiz-progress-text').innerText = 'Soal ' + (currentQIndex + 1) + ' dari ' + currentQuiz.length;
            document.getElementById('quiz-progress-bar').style.width = (((currentQIndex + 1) / currentQuiz.length) * 100) + '%';
            document.getElementById('quiz-score-live').innerText = 'Skor: ' + score;

            document.getElementById('quiz-ref-tag').innerText = q.reference_book || 'Dokumen Kurikulum Koding';
            document.getElementById('quiz-question-text').innerText = q.question;

            const optionsBox = document.getElementById('quiz-options-box');
            optionsBox.innerHTML = '';

            const explanationBox = document.getElementById('quiz-explanation');
            explanationBox.classList.add('hidden');

            const answered = userAnswers[currentQIndex] !== undefined;

            q.options.forEach((opt, idx) => {
                const btn = document.createElement('button');
                btn.className = "w-full text-left p-3.5 rounded-xl border text-xs font-medium transition flex items-center justify-between ";

                if (!answered) {
                    btn.className += "bg-zinc-950 border-zinc-800 text-zinc-300 hover:border-zinc-700 hover:bg-zinc-900";
                    btn.onclick = () => selectAnswer(idx);
                } else {
                    if (idx === q.correct_index) {
                        btn.className += "bg-emerald-950/40 border-emerald-500/50 text-emerald-300";
                    } else if (userAnswers[currentQIndex] === idx) {
                        btn.className += "bg-rose-950/40 border-rose-500/50 text-rose-300";
                    } else {
                        btn.className += "bg-zinc-950/50 border-zinc-800/50 text-zinc-600 cursor-not-allowed";
                    }
                }

                btn.innerHTML = '<div class="flex items-center gap-3">' +
                    '<span class="w-6 h-6 rounded-lg bg-zinc-800 flex items-center justify-center text-[10px] font-bold text-zinc-400">' + String.fromCharCode(65 + idx) + '</span>' +
                    '<span>' + escapeHtml(opt) + '</span></div>' +
                    (answered && idx === q.correct_index ? '<i class="fa-solid fa-circle-check text-emerald-400"></i>' : '') +
                    (answered && userAnswers[currentQIndex] === idx && idx !== q.correct_index ? '<i class="fa-solid fa-circle-xmark text-rose-400"></i>' : '');

                optionsBox.appendChild(btn);
            });

            if (answered) {
                explanationBox.classList.remove('hidden');
                const isCorrect = userAnswers[currentQIndex] === q.correct_index;
                const badge = document.getElementById('quiz-result-badge');
                badge.className = isCorrect ? "text-emerald-400 font-bold mb-1" : "text-rose-400 font-bold mb-1";
                badge.innerText = isCorrect ? "✓ Jawaban Kamu Benar!" : "✓ Jawaban Kamu Kurang Tepat!";
                document.getElementById('quiz-explanation-text').innerText = q.explanation;
            }

            document.getElementById('btn-prev-q').style.display = currentQIndex > 0 ? 'flex' : 'none';
            const btnNext = document.getElementById('btn-next-q');
            if (currentQIndex === currentQuiz.length - 1) {
                btnNext.innerHTML = '<span>Lihat Hasil Akhir</span> <i class="fa-solid fa-trophy"></i>';
            } else {
                btnNext.innerHTML = '<span>Soal Berikutnya</span> <i class="fa-solid fa-arrow-right"></i>';
            }
        }

        function selectAnswer(index) {
            if (userAnswers[currentQIndex] !== undefined) return;

            userAnswers[currentQIndex] = index;
            const q = currentQuiz[currentQIndex];
            if (index === q.correct_index) {
                score += 20;
            }
            renderQuestion();
        }

        function nextQuestion() {
            if (currentQIndex < currentQuiz.length - 1) {
                currentQIndex++;
                renderQuestion();
            } else {
                document.getElementById('quiz-container').classList.add('hidden');
                document.getElementById('quiz-final-result').classList.remove('hidden');
                document.getElementById('final-score-val').innerText = score + ' / 100';
                saveScore();
            }
        }

        async function saveScore() {
            if (!currentUser) return;
            const grade = document.getElementById('quiz-grade').value;
            try {
                const res = await fetch('/api/quiz/score/save', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        grade: grade,
                        score: score,
                        total_questions: currentQuiz.length
                    })
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    loadLeaderboard();
                }
            } catch (err) {
                console.error("Gagal menyimpan skor:", err);
            }
        }

        async function loadLeaderboard() {
            const tbody = document.getElementById('leaderboard-tbody');
            if (!tbody) return;
            try {
                const res = await fetch('/api/quiz/leaderboard');
                const data = await res.json();
                tbody.innerHTML = '';
                if (data.leaderboard && data.leaderboard.length > 0) {
                    data.leaderboard.forEach((entry, index) => {
                        const tr = document.createElement('tr');
                        tr.className = "hover:bg-zinc-800/30 transition";

                        let rankBadge = '';
                        if (index === 0) {
                            rankBadge = '<span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-amber-500/20 text-amber-400 border border-amber-500/30 font-bold text-xs"><i class="fa-solid fa-crown text-[10px]"></i></span>';
                        } else if (index === 1) {
                            rankBadge = '<span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-zinc-300/20 text-zinc-300 border border-zinc-400/30 font-bold text-xs">2</span>';
                        } else if (index === 2) {
                            rankBadge = '<span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-amber-700/20 text-amber-600 border border-amber-700/30 font-bold text-xs">3</span>';
                        } else {
                            rankBadge = '<span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-zinc-800 text-zinc-400 font-bold text-xs">' + (index + 1) + '</span>';
                        }

                        const displayName = escapeHtml(entry.full_name || entry.username);
                        const grade = escapeHtml(entry.high_grade || '-');

                        tr.innerHTML = '<td class="py-3 px-3">' + rankBadge + '</td>' +
                            '<td class="py-3 px-3 font-medium text-zinc-200">' + displayName + ' <span class="text-[10px] text-zinc-500">(@' + escapeHtml(entry.username) + ')</span></td>' +
                            '<td class="py-3 px-3 text-center"><span class="px-2 py-0.5 text-[10px] bg-zinc-800 text-zinc-400 rounded-md border border-zinc-700">' + grade + '</span></td>' +
                            '<td class="py-3 px-3 text-center text-zinc-300 font-mono">' + entry.quiz_count + '</td>' +
                            '<td class="py-3 px-3 text-right font-bold text-blue-400 font-mono">' + entry.total_score + '</td>';
                        tbody.appendChild(tr);
                    });
                } else {
                    tbody.innerHTML = '<tr><td colspan="5" class="py-6 text-center text-zinc-500">Belum ada data papan peringkat.</td></tr>';
                }
            } catch (err) {
                tbody.innerHTML = '<tr><td colspan="5" class="py-4 text-center text-red-400">Gagal memuat papan peringkat.</td></tr>';
            }
        }

        function prevQuestion() {
            if (currentQIndex > 0) {
                currentQIndex--;
                renderQuestion();
            }
        }

        // Chat Scripting
        userInput.addEventListener('input', function() {
            this.style.height = 'auto';
            this.style.height = (this.scrollHeight) + 'px';
        });

        userInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                chatForm.dispatchEvent(new Event('submit'));
            }
        });

        chatForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const text = userInput.value.trim();
            if (!text) return;

            appendMessage('user', text);
            userInput.value = '';
            userInput.style.height = 'auto';

            const loadingId = appendLoading();

            try {
                const res = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({prompt: text, session_id: activeSessionID})
                });

                const data = await res.json();
                removeLoading(loadingId);

                if (data.reply) {
                    if (data.session_id) {
                        activeSessionID = data.session_id;
                        if (currentUser) loadUserSessions();
                    }
                    if (typeof data.remaining_today !== 'undefined' && currentUser) {
                        currentUser.remaining_today = data.remaining_today;
                        currentUser.used_today = (currentUser.daily_limit || 0) - data.remaining_today;
                        renderUserNav();
                    }
                    appendMessage('assistant', data.reply);
                } else {
                    appendMessage('assistant', 'Maaf, terjadi kesalahan: ' + (data.error || 'Gagal memproses request.'));
                }
            } catch (err) {
                removeLoading(loadingId);
                appendMessage('assistant', 'Terjadi masalah jaringan ke server.');
            }
        });

        function appendMessage(role, text) {
            const div = document.createElement('div');
            div.className = 'flex items-start gap-3 ' + (role === 'user' ? 'flex-row-reverse' : '');

            const avatar = role === 'user' 
                ? '<div class="w-8 h-8 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 text-xs shrink-0"><i class="fa-solid fa-user"></i></div>'
                : '<div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xs shrink-0 shadow-md"><i class="fa-solid fa-sparkles"></i></div>';

            const contentHtml = role === 'assistant' ? marked.parse(text) : escapeHtml(text);
            const msgClass = role === 'user' ? 'aui-msg-user text-zinc-100' : 'aui-msg-assistant text-zinc-200 markdown-body';

            div.innerHTML = avatar +
                '<div class="' + msgClass + ' p-4 text-sm leading-relaxed max-w-xl shadow-sm">' + contentHtml + '</div>';

            chatBox.appendChild(div);
            chatBox.scrollTop = chatBox.scrollHeight;
        }

        function appendLoading() {
            const id = 'loading-' + Date.now();
            const div = document.createElement('div');
            div.id = id;
            div.className = 'flex items-start gap-3';
            div.innerHTML = '<div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xs shrink-0 shadow-md"><i class="fa-solid fa-sparkles"></i></div>' +
                '<div class="aui-msg-assistant p-4 text-sm text-zinc-400 flex items-center space-x-2">' +
                '<i class="fa-solid fa-spinner animate-spin text-blue-400"></i>' +
                '<span>Sedang mencari referensi halaman buku...</span></div>';
            chatBox.appendChild(div);
            chatBox.scrollTop = chatBox.scrollHeight;
            return id;
        }

        function removeLoading(id) {
            const el = document.getElementById(id);
            if (el) el.remove();
        }

        function escapeHtml(text) {
            return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
        }
    </script>
</body>
</html>`

	docs := knowledge.GetDocList()
	t, _ := template.New("home").Parse(tmpl)
	t.Execute(w, map[string]interface{}{"Docs": docs})
}

func HandleDocPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawFile := r.URL.Query().Get("file")
	mode := r.URL.Query().Get("mode") // "pdf" or "txt"
	if rawFile == "" {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}

	if mode == "txt" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"content": "Peninjauan teks TXT telah dinonaktifkan karena seluruh dokumen kini diakses langsung sebagai berkas PDF/DOCX asli."})
		return
	}

	filePath, mimeType, err := knowledge.GetOriginalFilePath(rawFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	// Serve raw original file (PDF / DOCX)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filepath.Base(filePath)+"\"")
	http.ServeFile(w, r, filePath)
}

func HandleGenerateQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var req struct {
		CategoryID int64  `json:"category_id"`
		Grade      string `json:"grade"`
		Topic      string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Grade = "Kelas 5 SD"
	}

	questions, err := service.GenerateQuizByCategory(req.CategoryID, req.Grade)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{"questions": questions})
	}
}

func HandleSaveQuizScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var req struct {
		Grade          string `json:"grade"`
		Score          int    `json:"score"`
		TotalQuestions int    `json:"total_questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Format request tidak valid"})
		return
	}

	if err := service.SaveQuizScore(user.ID, req.Grade, req.Score, req.TotalQuestions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard, err := service.GetLeaderboard()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if leaderboard == nil {
		leaderboard = []model.LeaderboardEntry{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"leaderboard": leaderboard})
}

func HandleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Harap masuk atau buat akun terlebih dahulu untuk melakukan obrolan dengan AI."})
		return
	}

	var req struct {
		Prompt    string `json:"prompt"`
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Pesan tidak boleh kosong."})
		return
	}

	// Check and update daily credit limit
	remainingToday, dailyLimit, err := auth.CheckAndUpdateDailyUsage(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":           err.Error(),
			"remaining_today": remainingToday,
			"daily_limit":     dailyLimit,
		})
		return
	}

	sessionID, reply, err := service.ProcessChatWithSession(user.ID, req.SessionID, req.Prompt)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"reply":           reply,
			"session_id":      sessionID,
			"remaining_today": remainingToday,
			"daily_limit":     dailyLimit,
		})
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Login - Koding Knowledge AI</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</head>
<body class="bg-black text-zinc-100 flex items-center justify-center min-h-screen font-sans p-4">
    <div class="w-full max-w-sm bg-zinc-900 border border-zinc-800 rounded-2xl p-6 shadow-2xl">
        <div class="flex flex-col items-center text-center mb-6">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white text-xl mb-3 shadow-lg shadow-blue-500/20">
                <i class="fa-solid fa-lock"></i>
            </div>
            <h1 class="text-lg font-bold text-zinc-100">Admin Portal Login</h1>
            <p class="text-xs text-zinc-400 mt-1">Masukkan password admin untuk mengelola API</p>
        </div>

        <form id="login-form" class="space-y-4">
            <div>
                <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">Password Admin</label>
                <input type="password" id="password" placeholder="Password Admin..." class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl px-4 py-2.5 text-sm focus:border-blue-500 focus:outline-none transition">
            </div>

            <button type="submit" class="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2.5 rounded-xl text-sm transition shadow-md flex items-center justify-center space-x-2">
                <span>Login Dashboard</span>
                <i class="fa-solid fa-right-to-bracket text-xs"></i>
            </button>
        </form>

        <div id="error-msg" class="mt-4 hidden p-3 rounded-xl bg-red-950/50 border border-red-800/50 text-red-400 text-xs text-center"></div>

        <div class="mt-6 text-center border-t border-zinc-800/80 pt-4">
            <a href="/" class="text-xs text-zinc-500 hover:text-zinc-300 transition">← Kembali ke Chatbot Public</a>
        </div>
    </div>

    document.getElementById('user-login-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;
        const errorMsg = document.getElementById('login-error-msg');

        try {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({username: username, password: password})
            });
            const data = await res.json();

            if (res.ok && data.token) {
                document.cookie = "user_token=" + data.token + "; path=/; max-age=2592000";
                if (data.user && data.user.role === 'admin') {
                    window.location.href = '/admin';
                } else {
                    window.location.href = '/';
                }
            } else {
                errorMsg.classList.remove('hidden');
                errorMsg.innerText = data.error || 'Login gagal';
            }
        } catch(err) {
            errorMsg.classList.remove('hidden');
            errorMsg.innerText = 'Terjadi kesalahan jaringan.';
        }
    });
    </script>
</body>
</html>`
	t, _ := template.New("login").Parse(tmpl)
	t.Execute(w, nil)
}

func HandleApiLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cfg := config.GetConfig()
	valid := (req.Password == cfg.AdminPassword)

	w.Header().Set("Content-Type", "application/json")
	if valid {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Password admin salah."})
	}
}

func HandleAdminUsersView(w http.ResponseWriter, r *http.Request) {
	u, err := auth.GetUserFromRequest(r)
	if err != nil || u == nil || u.Role != "admin" {
		http.Redirect(w, r, "/auth/login?redirect=/admin/users", http.StatusSeeOther)
		return
	}
	HandleAdminHTML(w, r, "users")
}

func HandleAdminPDFsView(w http.ResponseWriter, r *http.Request) {
	u, err := auth.GetUserFromRequest(r)
	if err != nil || u == nil || u.Role != "admin" {
		http.Redirect(w, r, "/auth/login?redirect=/admin/pdfs", http.StatusSeeOther)
		return
	}
	HandleAdminHTML(w, r, "pdfs")
}

func HandleAdminQuizzesView(w http.ResponseWriter, r *http.Request) {
	u, err := auth.GetUserFromRequest(r)
	if err != nil || u == nil || u.Role != "admin" {
		http.Redirect(w, r, "/auth/login?redirect=/admin/quizzes", http.StatusSeeOther)
		return
	}
	HandleAdminHTML(w, r, "quizzes")
}

func HandleAdminPackagesView(w http.ResponseWriter, r *http.Request) {
	u, err := auth.GetUserFromRequest(r)
	if err != nil || u == nil || u.Role != "admin" {
		http.Redirect(w, r, "/auth/login?redirect=/admin/packages", http.StatusSeeOther)
		return
	}
	HandleAdminHTML(w, r, "packages")
}

func HandleAdminConfigView(w http.ResponseWriter, r *http.Request) {
	u, err := auth.GetUserFromRequest(r)
	if err != nil || u == nil || u.Role != "admin" {
		http.Redirect(w, r, "/auth/login?redirect=/admin/config", http.StatusSeeOther)
		return
	}
	HandleAdminHTML(w, r, "config")
}

func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	HandleAdminUsersView(w, r)
}

func HandleAdminConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !verifyAdminAuth(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var newCfg model.AppConfig
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	currCfg := config.GetConfig()
	if newCfg.AdminPassword == "" {
		newCfg.AdminPassword = currCfg.AdminPassword
	}

	config.SaveConfig(newCfg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Admin API Handlers

func verifyAdminAuth(r *http.Request) bool {
	// Option 1: Check RBAC User Session Token (Role == "admin")
	u, err := auth.GetUserFromRequest(r)
	if err == nil && u != nil && u.Role == "admin" {
		return true
	}

	// Option 2: Fallback to Admin Password Header / Query Param
	adminPass := r.Header.Get("X-Admin-Password")
	if adminPass == "" {
		adminPass = r.URL.Query().Get("admin_password")
	}
	cfg := config.GetConfig()
	return adminPass != "" && adminPass == cfg.AdminPassword
}

func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}
	users, err := service.GetAllUsersAdmin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.AdminUserItem{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"users": users})
}

func HandleAdminUserDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.UserID == 0 {
		idStr := r.URL.Query().Get("user_id")
		req.UserID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	if req.UserID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User ID required"})
		return
	}

	if err := service.DeleteUserAdmin(req.UserID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminUserResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		UserID      int64  `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 || req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User ID and New Password required"})
		return
	}

	if err := service.ResetUserPasswordAdmin(req.UserID, req.NewPassword); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminPDFs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}
	pdfs, err := knowledge.GetDatasetPDFListAdmin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if pdfs == nil {
		pdfs = []model.AdminPDFItem{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"pdfs": pdfs})
}

func HandleAdminPDFUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	r.ParseMultipartForm(50 << 20) // 50MB max
	file, handler, err := r.FormFile("pdf_file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to read file: " + err.Error()})
		return
	}
	defer file.Close()

	fileName := filepath.Base(handler.Filename)
	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") && !strings.HasSuffix(strings.ToLower(fileName), ".docx") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Only .pdf and .docx files are supported"})
		return
	}

	destPath := filepath.Join(knowledge.DatasetDir, fileName)
	dst, err := os.Create(destPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to save file: " + err.Error()})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to write file contents"})
		return
	}

	// Reload knowledge base
	knowledge.LoadKnowledgeBase()

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": fileName})
}

func HandleAdminPDFDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Filename == "" {
		req.Filename = r.URL.Query().Get("filename")
	}

	if req.Filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Filename required"})
		return
	}

	if err := knowledge.DeleteDatasetPDFAdmin(req.Filename); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminQuizzes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	grade := r.URL.Query().Get("grade")
	topic := r.URL.Query().Get("topic")

	quizzes, err := service.GetCustomQuizzes(grade, topic)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if quizzes == nil {
		quizzes = []model.CustomQuizItem{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"quizzes": quizzes})
}

func HandleAdminQuizCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var q model.CustomQuizItem
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil || q.Question == "" || len(q.Options) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid quiz data"})
		return
	}

	if err := service.CreateCustomQuiz(q); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminQuizDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.ID == 0 {
		idStr := r.URL.Query().Get("id")
		req.ID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	if req.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Quiz ID required"})
		return
	}

	if err := service.DeleteCustomQuiz(req.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


// Auth & Session Handlers


// Quiz Categories Handlers

func HandleGetQuizCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cats, err := service.GetQuizCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if cats == nil {
		cats = []model.QuizCategoryItem{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"categories": cats})
}

func HandleAdminQuizCategoryCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var item model.QuizCategoryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil || item.Name == "" || item.Grade == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Nama dan Jenjang Kategori tidak boleh kosong"})
		return
	}

	if err := service.CreateQuizCategory(item); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminQuizCategoryDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "ID Kategori tidak valid"})
		return
	}

	if err := service.DeleteQuizCategory(req.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


func HandleAdminUserSetLimit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		UserID     int64 `json:"user_id"`
		DailyLimit int   `json:"daily_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 || req.DailyLimit < 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Data user atau limit kredit tidak valid"})
		return
	}

	if err := service.SetUserDailyLimitAdmin(req.UserID, req.DailyLimit); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


// Mayar.id Payment & Webhook Handlers

func HandleCreateQRISTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.MayarAPIKey) == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Gateway Pembayaran Mayar.id belum di-konfigurasi oleh Admin! Silakan hubungi Administrator."})
		return
	}

	user, err := auth.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Harap login terlebih dahulu untuk membeli paket."})
		return
	}

	var req struct {
		PackageID int64  `json:"package_id"`
		TierName  string `json:"tier_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Pilih paket kredit yang ingin dibeli."})
		return
	}

	tx, err := service.CreateMayarQRISTransaction(user.ID, req.PackageID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "transaction": tx})
}

func HandleGetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	txID := r.URL.Query().Get("tx_id")
	if txID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Transaction ID wajib diisi"})
		return
	}

	tx, err := service.GetPaymentTransactionStatus(txID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Transaksi tidak ditemukan"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"transaction": tx})
}

func HandleMayarWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Payload webhook tidak valid"})
		return
	}

	// Mayar Webhook Event payload extraction
	var txID string
	if dataObj, ok := body["data"].(map[string]interface{}); ok {
		if idVal, ok := dataObj["id"].(string); ok {
			txID = idVal
		}
		if txID == "" {
			if idVal, ok := dataObj["mobileVia"].(string); ok {
				txID = idVal
			}
		}
	}

	if txID != "" {
		_ = service.ProcessMayarPaymentSuccess(txID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


// Packages & Payment Handlers

func HandleGetCreditPackages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pkgs, err := service.GetCreditPackages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	cfg := config.GetConfig()
	isGatewayConfigured := strings.TrimSpace(cfg.MayarAPIKey) != ""

	json.NewEncoder(w).Encode(map[string]interface{}{
		"packages":             pkgs,
		"is_gateway_configured": isGatewayConfigured,
	})
}

func HandleAdminCreditPackages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}
	pkgs, err := service.GetAllCreditPackagesAdmin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"packages": pkgs})
}

func HandleAdminCreditPackageCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var pkg model.CreditPackageItem
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil || pkg.Name == "" || pkg.DailyLimit <= 0 || pkg.Price <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Nama, limit harian, dan harga paket wajib diisi valid."})
		return
	}

	if err := service.CreateCreditPackageAdmin(pkg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleAdminCreditPackageDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "ID paket tidak valid"})
		return
	}

	if err := service.DeleteCreditPackageAdmin(req.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


func HandleAdminCreditPackageUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !verifyAdminAuth(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized admin"})
		return
	}

	var pkg model.CreditPackageItem
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil || pkg.ID <= 0 || pkg.Name == "" || pkg.DailyLimit <= 0 || pkg.Price <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Data paket tidak valid."})
		return
	}

	if err := service.UpdateCreditPackageAdmin(pkg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


func HandleGetUserPaymentTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, err := auth.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	txs, err := service.GetUserPaymentTransactions(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if txs == nil {
		txs = []model.PaymentTransaction{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"transactions": txs})
}


func HandleCancelPaymentTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var req struct {
		TxID string `json:"tx_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TxID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "TxID wajib diisi"})
		return
	}

	if err := service.CancelPaymentTransactionUser(user.ID, req.TxID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func HandleDeletePaymentTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, err := auth.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var req struct {
		TxID string `json:"tx_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TxID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "TxID wajib diisi"})
		return
	}

	if err := service.DeletePaymentTransactionUser(user.ID, req.TxID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
