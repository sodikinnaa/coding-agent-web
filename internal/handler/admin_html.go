package handler

import (
	"html/template"
	"net/http"

	"coding_agent_web/internal/config"
)

func HandleAdminHTML(w http.ResponseWriter, r *http.Request, activeTab string) {
	tmpl := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Portal - AI Kurikulum Koding</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
        body { font-family: 'Inter', sans-serif; }
    </style>
</head>
<body class="bg-black text-zinc-100 flex flex-col min-h-screen">
    <!-- Header Admin -->
    <header class="h-16 border-b border-zinc-800 bg-zinc-950 px-4 md:px-6 flex justify-between items-center shrink-0">
        <div class="flex items-center space-x-2 md:space-x-3">
            <button onclick="toggleAdminMobileSidebar()" class="md:hidden text-zinc-400 hover:text-white p-1.5 rounded-lg border border-zinc-800 bg-zinc-900">
                <i class="fa-solid fa-bars text-sm"></i>
            </button>
            <div class="w-8 h-8 md:w-9 md:h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white font-bold text-sm md:text-base shadow-lg shadow-blue-500/20">
                <i class="fa-solid fa-user-shield"></i>
            </div>
            <div>
                <h1 class="text-xs md:text-sm font-bold text-white flex items-center gap-1.5">
                    <span>Admin Portal</span>
                    <span class="px-1.5 py-0.5 text-[9px] md:text-[10px] font-semibold bg-red-500/10 text-red-400 border border-red-500/20 rounded-full">Admin Only</span>
                </h1>
                <p class="text-[10px] md:text-[11px] text-zinc-400 hidden sm:block">Pengelolaan Pengguna, Berkas PDF, & Kuis Kurikulum</p>
            </div>
        </div>
        
        <div class="flex items-center space-x-2 md:space-x-3">
            <a href="/" class="text-[11px] md:text-xs bg-zinc-900 hover:bg-zinc-800 text-zinc-300 px-2.5 md:px-3.5 py-1.5 md:py-2 rounded-xl border border-zinc-800 transition flex items-center gap-1.5">
                <i class="fa-solid fa-house text-[10px] md:text-xs"></i>
                <span class="hidden sm:inline">Kembali ke Web Client</span>
                <span class="sm:hidden">Web</span>
            </a>
            <button onclick="adminLogout()" class="text-[11px] md:text-xs bg-red-500/10 hover:bg-red-500/20 text-red-400 px-2.5 md:px-3.5 py-1.5 md:py-2 rounded-xl border border-red-500/20 transition flex items-center gap-1.5">
                <i class="fa-solid fa-right-from-bracket text-[10px] md:text-xs"></i>
                <span class="hidden sm:inline">Logout Admin</span>
                <span class="sm:hidden">Keluar</span>
            </button>
        </div>
    </header>

    <div class="flex-1 flex overflow-hidden relative">
        <!-- Backdrop Mobile Sidebar -->
        <div id="admin-sidebar-backdrop" onclick="toggleAdminMobileSidebar()" class="fixed inset-0 bg-black/70 backdrop-blur-sm z-40 hidden md:hidden"></div>

        <!-- Sidebar Menu (Desktop & Mobile Drawer) -->
        <aside id="admin-sidebar" class="fixed md:static inset-y-0 left-0 z-50 w-64 bg-zinc-950 border-r border-zinc-800 p-4 flex flex-col gap-2 shrink-0 transform -translate-x-full md:translate-x-0 transition-transform duration-200">
            <div class="flex items-center justify-between px-3 py-1 md:hidden">
                <span class="text-xs font-bold text-white">Menu Pengelolaan</span>
                <button onclick="toggleAdminMobileSidebar()" class="text-zinc-400 hover:text-white">
                    <i class="fa-solid fa-xmark text-sm"></i>
                </button>
            </div>
            <div class="px-3 py-2 text-[10px] font-bold text-zinc-500 uppercase tracking-wider hidden md:block">Menu Pengelolaan</div>
            
            <button id="nav-btn-users" onclick="switchTab('users')" class="w-full text-left px-4 py-3 rounded-xl text-xs font-medium bg-blue-600 text-white transition flex items-center justify-between shadow-md">
                <div class="flex items-center space-x-2.5">
                    <i class="fa-solid fa-users text-sm"></i>
                    <span>Manajemen User</span>
                </div>
                <i class="fa-solid fa-chevron-right text-[10px]"></i>
            </button>

            <button id="nav-btn-pdfs" onclick="switchTab('pdfs')" class="w-full text-left px-4 py-3 rounded-xl text-xs font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition flex items-center justify-between">
                <div class="flex items-center space-x-2.5">
                    <i class="fa-solid fa-file-pdf text-sm"></i>
                    <span>Berkas Dataset PDF</span>
                </div>
                <i class="fa-solid fa-chevron-right text-[10px]"></i>
            </button>

            <button id="nav-btn-quizzes" onclick="switchTab('quizzes')" class="w-full text-left px-4 py-3 rounded-xl text-xs font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition flex items-center justify-between">
                <div class="flex items-center space-x-2.5">
                    <i class="fa-solid fa-layer-group text-sm"></i>
                    <span>Kategori & Soal Kuis</span>
                </div>
                <i class="fa-solid fa-chevron-right text-[10px]"></i>
            </button>

            <button id="nav-btn-packages" onclick="switchTab('packages')" class="w-full text-left px-4 py-3 rounded-xl text-xs font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition flex items-center justify-between">
                <div class="flex items-center space-x-2.5">
                    <i class="fa-solid fa-boxes-packing text-sm text-emerald-400"></i>
                    <span>Paket Kredit QRIS</span>
                </div>
                <i class="fa-solid fa-chevron-right text-[10px]"></i>
            </button>

            <button id="nav-btn-config" onclick="switchTab('config')" class="w-full text-left px-4 py-3 rounded-xl text-xs font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition flex items-center justify-between">
                <div class="flex items-center space-x-2.5">
                    <i class="fa-solid fa-sliders text-sm"></i>
                    <span>Konfigurasi API AI</span>
                </div>
                <i class="fa-solid fa-chevron-right text-[10px]"></i>
            </button>
        </aside>

        <!-- Main Workspace -->
        <main class="flex-1 bg-black p-6 overflow-y-auto">
            <!-- TAB 1: MANAJEMEN USER -->
            <section id="tab-users" class="space-y-6">
                <div class="flex justify-between items-center bg-zinc-900 border border-zinc-800 p-5 rounded-2xl">
                    <div>
                        <h2 class="text-base font-bold text-white flex items-center gap-2">
                            <i class="fa-solid fa-users text-blue-500"></i>
                            <span>Daftar Pengguna Terdaftar</span>
                        </h2>
                        <p class="text-xs text-zinc-400 mt-1">Kelola data pengguna, role akses, reset password, & hapus akun</p>
                    </div>
                    <button onclick="loadUsers()" class="px-3.5 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-xl border border-zinc-700 transition flex items-center gap-1.5">
                        <i class="fa-solid fa-rotate-right"></i>
                        <span>Refresh Data</span>
                    </button>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-2xl overflow-x-auto shadow-xl">
                    <table class="w-full text-left text-xs min-w-[700px]">
                        <thead class="bg-zinc-950 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800">
                            <tr>
                                <th class="p-4">ID</th>
                                <th class="p-4">Username</th>
                                <th class="p-4">Nama Lengkap</th>
                                <th class="p-4">Role</th>
                                <th class="p-4">Limit Chat / Hari</th>
                                <th class="p-4">Tanggal Daftar</th>
                                <th class="p-4 text-center">Total Poin</th>
                                <th class="p-4 text-right">Aksi</th>
                            </tr>
                        </thead>
                        <tbody id="users-table-body" class="divide-y divide-zinc-800/60 text-zinc-300">
                            <tr><td colspan="8" class="p-6 text-center text-zinc-500">Memuat data pengguna...</td></tr>
                        </tbody>
                    </table>
                </div>
            </section>

            <!-- TAB 2: DATASET PDF -->
            <section id="tab-pdfs" class="space-y-6 hidden">
                <div class="flex justify-between items-center bg-zinc-900 border border-zinc-800 p-5 rounded-2xl">
                    <div>
                        <h2 class="text-base font-bold text-white flex items-center gap-2">
                            <i class="fa-solid fa-file-pdf text-red-500"></i>
                            <span>Manajemen Berkas Dataset PDF/DOCX</span>
                        </h2>
                        <p class="text-xs text-zinc-400 mt-1">File dokumen asli di <code>/root/coding_dataset</code> yang dibaca langsung oleh Direct PDF Vision AI</p>
                    </div>
                    <button onclick="openUploadPDFModal()" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl transition flex items-center gap-2 shadow-lg shadow-blue-500/20">
                        <i class="fa-solid fa-cloud-arrow-up"></i>
                        <span>Upload Berkas PDF Baru</span>
                    </button>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-2xl overflow-x-auto shadow-xl">
                    <table class="w-full text-left text-xs min-w-[700px]">
                        <thead class="bg-zinc-950 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800">
                            <tr>
                                <th class="p-4">Nama Berkas</th>
                                <th class="p-4">Ukuran File</th>
                                <th class="p-4">Waktu Modifikasi</th>
                                <th class="p-4 text-right">Aksi</th>
                            </tr>
                        </thead>
                        <tbody id="pdfs-table-body" class="divide-y divide-zinc-800/60 text-zinc-300">
                            <tr><td colspan="4" class="p-6 text-center text-zinc-500">Memuat data berkas PDF...</td></tr>
                        </tbody>
                    </table>
                </div>
            </section>

            <!-- TAB 3: KUIS CATEGORIES -->
            <section id="tab-quizzes" class="space-y-6 hidden">
                <div class="flex justify-between items-center bg-zinc-900 border border-zinc-800 p-5 rounded-2xl">
                    <div>
                        <h2 class="text-base font-bold text-white flex items-center gap-2">
                            <i class="fa-solid fa-layer-group text-indigo-500"></i>
                            <span>Pengelolaan Kategori Kuis AI</span>
                        </h2>
                        <p class="text-xs text-zinc-400 mt-1">Buat Kategori Kuis custom, pilih Buku Referensi Dataset PDF, & sesuaikan Jumlah Soal AI</p>
                    </div>
                    <button onclick="openCreateCategoryModal()" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl transition flex items-center gap-2 shadow-lg shadow-indigo-500/20">
                        <i class="fa-solid fa-plus"></i>
                        <span>Buat Kategori Kuis Baru</span>
                    </button>
                </div>

                <div id="quizzes-container" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <!-- Dynamic Quiz Category Cards -->
                </div>
            </section>

            <!-- TAB 4: PACKAGES -->
            <section id="tab-packages" class="space-y-6 hidden">
                <div class="flex justify-between items-center bg-zinc-900 border border-zinc-800 p-5 rounded-2xl">
                    <div>
                        <h2 class="text-base font-bold text-white flex items-center gap-2">
                            <i class="fa-solid fa-boxes-packing text-emerald-500"></i>
                            <span>Pengelolaan Paket Kredit QRIS (Mayar.id)</span>
                        </h2>
                        <p class="text-xs text-zinc-400 mt-1">Buat, edit, atau atur daftar paket kredit chat harian yang dijual ke pengguna</p>
                    </div>
                    <button onclick="openCreatePackageModal()" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold rounded-xl transition flex items-center gap-2 shadow-lg shadow-emerald-500/20">
                        <i class="fa-solid fa-plus"></i>
                        <span>Buat Paket Baru</span>
                    </button>
                </div>

                <div class="bg-zinc-900 border border-zinc-800 rounded-2xl overflow-x-auto shadow-xl">
                    <table class="w-full text-left text-xs min-w-[700px]">
                        <thead class="bg-zinc-950 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800">
                            <tr>
                                <th class="p-4">ID</th>
                                <th class="p-4">Nama Paket</th>
                                <th class="p-4 text-center">Limit Chat / Hari</th>
                                <th class="p-4">Harga (Rp)</th>
                                <th class="p-4">Deskripsi</th>
                                <th class="p-4 text-right">Aksi</th>
                            </tr>
                        </thead>
                        <tbody id="packages-table-body" class="divide-y divide-zinc-800/60 text-zinc-300">
                            <tr><td colspan="6" class="p-6 text-center text-zinc-500">Memuat data paket kredit...</td></tr>
                        </tbody>
                    </table>
                </div>
            </section>

            <!-- TAB 5: CONFIG AI -->
            <section id="tab-config" class="space-y-6 hidden">
                <div class="bg-zinc-900 border border-zinc-800 p-6 rounded-2xl max-w-2xl mx-auto shadow-2xl space-y-5">
                    <div class="flex items-center space-x-3 pb-4 border-b border-zinc-800">
                        <div class="w-10 h-10 rounded-xl bg-blue-600/20 border border-blue-500/30 text-blue-400 flex items-center justify-center text-lg">
                            <i class="fa-solid fa-sliders"></i>
                        </div>
                        <div>
                            <h2 class="text-base font-bold text-white">Konfigurasi AI Backend API</h2>
                            <p class="text-xs text-zinc-400">Pengaturan Endpoint OpenAI-Compatible & Provider AI Model</p>
                        </div>
                    </div>

                    <form id="config-form" class="space-y-4">
                        <div>
                            <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">Base URL Endpoint API</label>
                            <input type="text" id="cfg-base-url" value="{{.Config.BaseURL}}" required class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl px-4 py-2.5 text-xs font-mono focus:border-blue-500 focus:outline-none">
                        </div>

                        <div>
                            <div class="flex items-center justify-between mb-2">
                                <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider">API Key Authorization</label>
                                <span class="px-2 py-0.5 text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full font-medium flex items-center gap-1">
                                    <i class="fa-solid fa-shield-halved text-[10px]"></i> AES-256-GCM Encrypted
                                </span>
                            </div>
                            <div class="relative flex items-center">
                                <input type="password" id="cfg-api-key" value="{{.Config.APIKey}}" required class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl pl-4 pr-10 py-2.5 text-xs font-mono focus:border-blue-500 focus:outline-none">
                                <button type="button" onclick="togglePasswordVisibility('cfg-api-key', 'eye-icon-api-key')" class="absolute right-3 text-zinc-400 hover:text-white transition p-1" title="Lihat/Sembunyikan API Key">
                                    <i id="eye-icon-api-key" class="fa-solid fa-eye-slash text-xs"></i>
                                </button>
                            </div>
                        </div>

                        <div>
                            <div class="flex items-center justify-between mb-2">
                                <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider">Model AI Name</label>
                                <button type="button" onclick="fetchVisionModels()" id="btn-fetch-models" class="px-2.5 py-1 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-500/30 rounded-lg text-[11px] font-medium transition flex items-center gap-1.5 active:scale-95" title="Fetch & filter vision models from Provider API">
                                    <i class="fa-solid fa-arrows-rotate text-[10px]" id="fetch-model-icon"></i>
                                    <span>Fetch Vision Models</span>
                                </button>
                            </div>
                            <div class="space-y-2">
                                <div class="flex items-center space-x-2">
                                    <input type="text" id="cfg-model" value="{{.Config.Model}}" required placeholder="Tulis atau pilih model AI..." class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl px-4 py-2.5 text-xs font-mono focus:border-blue-500 focus:outline-none">
                                    <select id="cfg-model-select" onchange="if(this.value) document.getElementById('cfg-model').value = this.value" class="bg-zinc-950 border border-zinc-800 text-zinc-300 rounded-xl px-3 py-2.5 text-xs font-mono focus:border-blue-500 focus:outline-none max-w-[200px]">
                                        <option value="">-- Pilih Model Provider --</option>
                                        <option value="{{.Config.Model}}" selected>{{.Config.Model}}</option>
                                    </select>
                                </div>
                                <p class="text-[10px] text-zinc-500 flex items-center gap-1">
                                    <i class="fa-solid fa-circle-info text-blue-400"></i>
                                    <span>Tekan <strong>Fetch Vision Models</strong> untuk mengambil daftar model multimodal/vision aktif dari Provider API secara otomatis.</span>
                                </p>
                            </div>
                        </div>

                        <div>
                            <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">System Prompt AI</label>
                            <textarea id="cfg-system-prompt" rows="3" required class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl px-4 py-2.5 text-xs focus:border-blue-500 focus:outline-none">{{.Config.SystemPrompt}}</textarea>
                        </div>

                        <div class="pt-4 border-t border-zinc-800/80 space-y-3">
                            <div class="flex items-center space-x-2 text-emerald-400 font-bold text-xs">
                                <i class="fa-solid fa-qrcode text-sm"></i>
                                <span>Mayar.id Dynamic QRIS Payment Integration</span>
                            </div>
                            <div>
                                <div class="flex items-center justify-between mb-2">
                                    <label class="block text-xs font-semibold text-zinc-400 uppercase tracking-wider">Mayar.id API Key Authorization</label>
                                    <span class="px-2 py-0.5 text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full font-medium flex items-center gap-1">
                                        <i class="fa-solid fa-shield-halved text-[10px]"></i> AES-256-GCM Encrypted
                                    </span>
                                </div>
                                <div class="relative flex items-center">
                                    <input type="password" id="cfg-mayar-key" value="{{.Config.MayarAPIKey}}" placeholder="mayar_sec_key_..." class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl pl-4 pr-10 py-2.5 text-xs font-mono focus:border-emerald-500 focus:outline-none">
                                    <button type="button" onclick="togglePasswordVisibility('cfg-mayar-key', 'eye-icon-mayar-key')" class="absolute right-3 text-zinc-400 hover:text-white transition p-1" title="Lihat/Sembunyikan Mayar Key">
                                        <i id="eye-icon-mayar-key" class="fa-solid fa-eye-slash text-xs"></i>
                                    </button>
                                </div>
                            </div>
                            <div class="bg-zinc-950 border border-zinc-800 p-3.5 rounded-xl space-y-2 text-xs">
                                <label class="block text-[11px] font-semibold text-zinc-400">Webhook Notification URL (Setup Wizard):</label>
                                <div class="flex items-center space-x-2">
                                    <input type="text" readonly value="https://aikurikulum.siapdigital.my.id/api/mayar/webhook" class="w-full bg-zinc-900 border border-zinc-800 text-emerald-400 font-mono text-[11px] rounded-lg px-3 py-1.5 focus:outline-none">
                                </div>
                                <p class="text-[10px] text-zinc-500 leading-relaxed">Masukkan URL Webhook di atas pada dashboard Mayar.id -> Integration -> Webhooks agar pembayaran QRIS ter-konfirmasi & kuota ter-upgrade secara otomatis.</p>
                            </div>
                        </div>

                        <div class="pt-3">
                            <button type="submit" class="w-full bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold py-3 rounded-xl transition shadow-lg shadow-blue-500/20 flex items-center justify-center gap-2">
                                <i class="fa-solid fa-floppy-disk"></i>
                                <span>Simpan Konfigurasi AI</span>
                            </button>
                        </div>
                    </form>
                </div>
            </section>
        </main>
    </div>

    
    <!-- Modal Set Limit Paket -->
    <div id="set-limit-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
                <i class="fa-solid fa-bolt text-emerald-400"></i>
                <span>Atur Paket Kredit Harian User</span>
            </h3>
            <p id="limit-target-name" class="text-xs text-zinc-400"></p>
            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1.5">Pilih Paket Daily Reset</label>
                <select id="limit-tier-select" onchange="toggleCustomLimitInput()" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2.5 text-zinc-100 focus:outline-none">
                    <option value="5">Free Tier (5 Chat / Day)</option>
                    <option value="50">Bronze Tier (50 Chat / Day)</option>
                    <option value="100">Silver Tier (100 Chat / Day)</option>
                    <option value="200">Gold Tier (200 Chat / Day)</option>
                    <option value="custom">Custom Admin Limit...</option>
                </select>
            </div>
            <div id="custom-limit-box" class="hidden">
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Jumlah Kredit Custom per Hari</label>
                <input type="number" id="custom-limit-val" value="500" min="1" class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl px-4 py-2 text-xs focus:outline-none">
            </div>
            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeSetLimitModal()" class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-xl transition">Batal</button>
                <button onclick="submitSetLimit()" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold rounded-xl transition">Simpan Paket</button>
            </div>
        </div>
    </div>

    <!-- Modal Reset Password -->
    <div id="reset-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
                <i class="fa-solid fa-key text-blue-500"></i>
                <span>Reset Password Pengguna</span>
            </h3>
            <p id="reset-target-name" class="text-xs text-zinc-400"></p>
            <div class="relative flex items-center">
                <input type="password" id="reset-new-pass" placeholder="Password Baru..." class="w-full bg-zinc-950 border border-zinc-800 text-zinc-100 rounded-xl pl-4 pr-10 py-2.5 text-xs focus:border-blue-500 focus:outline-none">
                <button type="button" onclick="togglePasswordVisibility('reset-new-pass', 'eye-icon-reset')" class="absolute right-3 text-zinc-400 hover:text-white transition p-1">
                    <i id="eye-icon-reset" class="fa-solid fa-eye text-xs"></i>
                </button>
            </div>
            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeResetModal()" class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-xl transition">Batal</button>
                <button onclick="submitResetPassword()" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl transition">Simpan Password</button>
            </div>
        </div>
    </div>

    <!-- Modal Upload PDF -->
    <div id="upload-pdf-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
                <i class="fa-solid fa-cloud-arrow-up text-blue-500"></i>
                <span>Upload Berkas PDF/DOCX Baru</span>
            </h3>
            <p class="text-xs text-zinc-400">Berkas akan langsung masuk ke dataset /root/coding_dataset</p>
            <input type="file" id="pdf-file-input" accept=".pdf,.docx" class="w-full text-xs text-zinc-300 file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-blue-600 file:text-white hover:file:bg-blue-500">
            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeUploadPDFModal()" class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-xl transition">Batal</button>
                <button onclick="submitUploadPDF()" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl transition">Upload File</button>
            </div>
        </div>
    </div>

    
    <!-- Modal Create / Edit Package -->
    <div id="create-package-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
                <i id="pkg-modal-icon" class="fa-solid fa-plus text-emerald-400"></i>
                <span id="pkg-modal-title">Buat Paket Kredit Baru</span>
            </h3>
            
            <input type="hidden" id="edit-pkg-id" value="0">

            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Nama Paket</label>
                <input type="text" id="new-pkg-name" placeholder="Contoh: Platinum Tier" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
            </div>

            <div class="grid grid-cols-2 gap-3">
                <div>
                    <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Limit Chat / Hari</label>
                    <input type="number" id="new-pkg-limit" value="500" min="1" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
                </div>
                <div>
                    <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Harga (Rp)</label>
                    <input type="number" id="new-pkg-price" value="75000" min="1000" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
                </div>
            </div>

            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Deskripsi Paket</label>
                <input type="text" id="new-pkg-desc" placeholder="Contoh: 500 Chat per Hari (Daily Reset)" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
            </div>

            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeCreatePackageModal()" class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-xl transition">Batal</button>
                <button onclick="submitSavePackage()" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold rounded-xl transition">Simpan Paket</button>
            </div>
        </div>
    </div>

    <!-- Modal Create Custom Quiz Category -->
    <div id="create-quiz-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 hidden">
        <div class="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-lg p-6 shadow-2xl space-y-4 overflow-y-auto max-h-[90vh]">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
                <i class="fa-solid fa-plus text-indigo-500"></i>
                <span>Buat Kategori Kuis AI Baru</span>
            </h3>
            
            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Nama Kategori Kuis</label>
                <input type="text" id="new-cat-name" placeholder="Contoh: Evaluasi Etika & Algoritma SD" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
            </div>

            <div class="grid grid-cols-2 gap-3">
                <div>
                    <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Jenjang Sekolah</label>
                    <select id="new-cat-grade" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
                        <option value="Kelas 5 SD">Kelas 5 SD</option>
                        <option value="Kelas 6 SD">Kelas 6 SD</option>
                        <option value="Kelas 7 SMP">Kelas 7 SMP</option>
                        <option value="Kelas 10 SMA">Kelas 10 SMA</option>
                    </select>
                </div>
                <div>
                    <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Jumlah Soal AI</label>
                    <input type="number" id="new-cat-questions" value="5" min="1" max="20" class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl px-3 py-2 text-zinc-100 focus:outline-none">
                </div>
            </div>

            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Pilih Berkas Buku Referensi (Multi-select)</label>
                <div id="cat-books-checkboxes" class="bg-zinc-950 border border-zinc-800 p-3 rounded-xl max-h-36 overflow-y-auto space-y-1.5 text-xs">
                    <!-- Dynamic PDF Checkboxes -->
                </div>
            </div>

            <div>
                <label class="block text-[11px] font-semibold text-zinc-400 mb-1">Deskripsi Kuis (Opsional)</label>
                <textarea id="new-cat-desc" rows="2" placeholder="Penjelasan singkat mengenai kuis..." class="w-full bg-zinc-950 border border-zinc-800 text-xs rounded-xl p-2.5 text-zinc-100 focus:outline-none"></textarea>
            </div>

            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeCreateQuizModal()" class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-xl transition">Batal</button>
                <button onclick="submitCreateCategory()" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl transition">Simpan Kategori Kuis</button>
            </div>
        </div>
    </div>

    <script>
        function toggleAdminMobileSidebar() {
            const sidebar = document.getElementById('admin-sidebar');
            const backdrop = document.getElementById('admin-sidebar-backdrop');
            if (!sidebar || !backdrop) return;

            if (sidebar.classList.contains('-translate-x-full')) {
                sidebar.classList.remove('-translate-x-full');
                backdrop.classList.remove('hidden');
            } else {
                sidebar.classList.add('-translate-x-full');
                backdrop.classList.add('hidden');
            }
        }

        function togglePasswordVisibility(inputId, iconId) {
            const input = document.getElementById(inputId);
            const icon = document.getElementById(iconId);
            if (!input || !icon) return;

            if (input.type === 'password') {
                input.type = 'text';
                icon.className = 'fa-solid fa-eye-slash text-xs';
            } else {
                input.type = 'password';
                icon.className = 'fa-solid fa-eye text-xs';
            }
        }

        function escapeHtml(str) {
            if (!str) return '';
            return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
        }

        function adminLogout() {
            document.cookie = "user_token=; path=/; max-age=-1";
            window.location.href = "/";
        }

        
        // --- TAB 5: ADMIN PACKAGES ---
        let adminPackagesCache = [];

        async function loadAdminPackages() {
            const tbody = document.getElementById('packages-table-body');
            if (!tbody) return;
            tbody.innerHTML = '<tr><td colspan="6" class="p-6 text-center text-zinc-500">Memuat paket kredit...</td></tr>';
            try {
                const res = await fetch('/api/admin/packages');
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Gagal memuat paket');

                if (!data.packages || data.packages.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="6" class="p-6 text-center text-zinc-500">Belum ada paket kredit yang dibuat.</td></tr>';
                    return;
                }

                adminPackagesCache = data.packages;
                let html = '';
                data.packages.forEach(p => {
                    html += '<tr class="hover:bg-zinc-800/40">' +
                        '<td class="p-4 font-mono font-bold">' + p.id + '</td>' +
                        '<td class="p-4 font-medium text-white">' + escapeHtml(p.name) + '</td>' +
                        '<td class="p-4 text-center"><span class="px-2.5 py-0.5 text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full">' + p.daily_limit + ' Chat/Day</span></td>' +
                        '<td class="p-4 font-bold text-emerald-400">Rp ' + p.price.toLocaleString('id-ID') + '</td>' +
                        '<td class="p-4 text-zinc-400">' + escapeHtml(p.description || '-') + '</td>' +
                        '<td class="p-4 text-right space-x-1.5">' +
                            '<button onclick="openEditPackageModal(' + p.id + ')" class="px-2.5 py-1 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border border-blue-500/20 rounded-lg text-[11px] transition"><i class="fa-solid fa-pen-to-square text-[10px]"></i> Edit</button>' +
                            '<button onclick="deletePackage(' + p.id + ')" class="px-2.5 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-[11px] transition"><i class="fa-solid fa-trash text-[10px]"></i> Hapus</button>' +
                        '</td>' +
                    '</tr>';
                });
                tbody.innerHTML = html;
            } catch(e) {
                tbody.innerHTML = '<tr><td colspan="6" class="p-6 text-center text-red-400">Error: ' + escapeHtml(e.message) + '</td></tr>';
            }
        }

        function openCreatePackageModal() {
            document.getElementById('edit-pkg-id').value = "0";
            document.getElementById('pkg-modal-title').innerText = "Buat Paket Kredit Baru";
            document.getElementById('pkg-modal-icon').className = "fa-solid fa-plus text-emerald-400";
            document.getElementById('new-pkg-name').value = "";
            document.getElementById('new-pkg-limit').value = "500";
            document.getElementById('new-pkg-price').value = "75000";
            document.getElementById('new-pkg-desc').value = "";
            document.getElementById('create-package-modal').classList.remove('hidden');
        }

        function openEditPackageModal(id) {
            const pkg = adminPackagesCache.find(p => p.id === id);
            if (!pkg) return;
            document.getElementById('edit-pkg-id').value = String(pkg.id);
            document.getElementById('pkg-modal-title').innerText = "Edit Paket Kredit #" + pkg.id;
            document.getElementById('pkg-modal-icon').className = "fa-solid fa-pen-to-square text-blue-400";
            document.getElementById('new-pkg-name').value = pkg.name;
            document.getElementById('new-pkg-limit').value = pkg.daily_limit;
            document.getElementById('new-pkg-price').value = pkg.price;
            document.getElementById('new-pkg-desc').value = pkg.description || "";
            document.getElementById('create-package-modal').classList.remove('hidden');
        }

        function closeCreatePackageModal() {
            document.getElementById('create-package-modal').classList.add('hidden');
        }

        async function submitSavePackage() {
            const editId = parseInt(document.getElementById('edit-pkg-id').value) || 0;
            const name = document.getElementById('new-pkg-name').value.trim();
            const limit = parseInt(document.getElementById('new-pkg-limit').value) || 50;
            const price = parseInt(document.getElementById('new-pkg-price').value) || 15000;
            const desc = document.getElementById('new-pkg-desc').value.trim();

            if (!name) return alert('Nama Paket wajib diisi!');

            const endpoint = editId > 0 ? '/api/admin/package/update' : '/api/admin/package/create';
            const payload = editId > 0 
                ? { id: editId, name: name, daily_limit: limit, price: price, description: desc }
                : { name: name, daily_limit: limit, price: price, description: desc };

            try {
                const res = await fetch(endpoint, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(payload)
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Paket Kredit berhasil disimpan!');
                    closeCreatePackageModal();
                    loadAdminPackages();
                } else {
                    alert('Gagal menyimpan paket: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server.');
            }
        }

        async function deletePackage(id) {
            if (!confirm('Apakah Anda yakin ingin menghapus paket kredit ini?')) return;
            try {
                const res = await fetch('/api/admin/package/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({id: id})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Paket Kredit berhasil dihapus!');
                    loadAdminPackages();
                } else {
                    alert('Gagal hapus paket: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server.');
            }
        }

        function switchTab(tab) {
            ['users', 'pdfs', 'quizzes', 'packages', 'config'].forEach(t => {
                const sec = document.getElementById('tab-' + t);
                const btn = document.getElementById('nav-btn-' + t);
                if (sec) sec.classList.add('hidden');
                if (btn) {
                    btn.className = "w-full text-left px-4 py-3 rounded-xl text-xs font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200 transition flex items-center justify-between";
                }
            });

            const activeSec = document.getElementById('tab-' + tab);
            const activeBtn = document.getElementById('nav-btn-' + tab);
            if (activeSec) activeSec.classList.remove('hidden');
            if (activeBtn) {
                activeBtn.className = "w-full text-left px-4 py-3 rounded-xl text-xs font-medium bg-blue-600 text-white transition flex items-center justify-between shadow-md";
            }

            // Close mobile sidebar after tab switch
            if (window.innerWidth < 768) {
                const sidebar = document.getElementById('admin-sidebar');
                const backdrop = document.getElementById('admin-sidebar-backdrop');
                if (sidebar) sidebar.classList.add('-translate-x-full');
                if (backdrop) backdrop.classList.add('hidden');
            }
            if (tab === 'users') { history.pushState(null, '', '/admin/users'); loadUsers(); }
            if (tab === 'pdfs') { history.pushState(null, '', '/admin/pdfs'); loadPDFs(); }
            if (tab === 'quizzes') { history.pushState(null, '', '/admin/quizzes'); loadQuizzes(); }
            if (tab === 'packages') { history.pushState(null, '', '/admin/packages'); loadAdminPackages(); }
            if (tab === 'config') { history.pushState(null, '', '/admin/config'); }
        }

        // --- TAB 1: USERS ---
        async function loadUsers() {
            const tbody = document.getElementById('users-table-body');
            tbody.innerHTML = '<tr><td colspan="8" class="p-6 text-center text-zinc-500">Memuat data pengguna...</td></tr>';
            try {
                const res = await fetch('/api/admin/users');
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Gagal memuat user');

                if (!data.users || data.users.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="8" class="p-6 text-center text-zinc-500">Belum ada pengguna terdaftar.</td></tr>';
                    return;
                }

                let html = '';
                data.users.forEach(u => {
                    let roleBadge = u.role === 'admin' 
                        ? '<span class="px-2 py-0.5 text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20 rounded-full">ADMIN</span>'
                        : '<span class="px-2 py-0.5 text-[10px] font-medium bg-zinc-800 text-zinc-400 rounded-full">USER</span>';

                    let limitBadge = u.role === 'admin' 
                        ? '<span class="px-2 py-0.5 text-[10px] font-bold bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded-full">UNLIMITED</span>'
                        : '<span class="px-2.5 py-0.5 text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full">' + (u.used_today || 0) + ' / ' + (u.daily_limit || 5) + ' Chat/Day</span>';

                    let formattedDate = new Date(u.created_at).toLocaleDateString('id-ID');

                    html += '<tr class="hover:bg-zinc-800/40">' +
                        '<td class="p-4 font-mono font-bold">' + u.id + '</td>' +
                        '<td class="p-4 font-medium text-white">' + escapeHtml(u.username) + '</td>' +
                        '<td class="p-4">' + escapeHtml(u.full_name || '-') + '</td>' +
                        '<td class="p-4">' + roleBadge + '</td>' +
                        '<td class="p-4">' + limitBadge + '</td>' +
                        '<td class="p-4 text-zinc-400">' + formattedDate + '</td>' +
                        '<td class="p-4 text-center font-bold text-amber-400">' + (u.total_score || 0) + '</td>' +
                        '<td class="p-4 text-right space-x-1.5">' +
                            '<button onclick="openSetLimitModal(' + u.id + ', \'' + escapeHtml(u.username) + '\', ' + u.daily_limit + ')" class="px-2.5 py-1 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/20 rounded-lg text-[11px] transition"><i class="fa-solid fa-bolt text-[10px]"></i> Paket</button>' +
                            '<button onclick="openResetModal(' + u.id + ', \'' + escapeHtml(u.username) + '\')" class="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700 rounded-lg text-[11px] transition"><i class="fa-solid fa-key text-[10px]"></i> Pass</button>' +
                            (u.role !== 'admin' ? '<button onclick="deleteUser(' + u.id + ', \'' + escapeHtml(u.username) + '\')" class="px-2.5 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-[11px] transition"><i class="fa-solid fa-trash text-[10px]"></i> Hapus</button>' : '') +
                        '</td>' +
                    '</tr>';
                });
                tbody.innerHTML = html;
            } catch (err) {
                tbody.innerHTML = '<tr><td colspan="8" class="p-6 text-center text-red-400">Error: ' + escapeHtml(err.message) + '</td></tr>';
            }
        }

        let selectedResetUserID = null;
        
        let selectedLimitUserID = null;

        function openSetLimitModal(id, username, currentLimit) {
            selectedLimitUserID = id;
            document.getElementById('limit-target-name').innerText = 'Mengubah paket harian untuk pengguna: ' + username;
            const select = document.getElementById('limit-tier-select');
            if ([5, 50, 100, 200].includes(currentLimit)) {
                select.value = String(currentLimit);
                document.getElementById('custom-limit-box').classList.add('hidden');
            } else {
                select.value = 'custom';
                document.getElementById('custom-limit-box').classList.remove('hidden');
                document.getElementById('custom-limit-val').value = currentLimit;
            }
            document.getElementById('set-limit-modal').classList.remove('hidden');
        }

        function toggleCustomLimitInput() {
            const select = document.getElementById('limit-tier-select');
            const customBox = document.getElementById('custom-limit-box');
            if (select.value === 'custom') {
                customBox.classList.remove('hidden');
            } else {
                customBox.classList.add('hidden');
            }
        }

        function closeSetLimitModal() {
            document.getElementById('set-limit-modal').classList.add('hidden');
        }

        async function submitSetLimit() {
            const selectVal = document.getElementById('limit-tier-select').value;
            let finalLimit = 5;
            if (selectVal === 'custom') {
                finalLimit = parseInt(document.getElementById('custom-limit-val').value) || 5;
            } else {
                finalLimit = parseInt(selectVal);
            }

            try {
                const res = await fetch('/api/admin/user/set-limit', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({user_id: selectedLimitUserID, daily_limit: finalLimit})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Paket kredit harian pengguna berhasil diperbarui!');
                    closeSetLimitModal();
                    loadUsers();
                } else {
                    alert('Gagal update limit: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        function openResetModal(id, username) {
            selectedResetUserID = id;
            document.getElementById('reset-target-name').innerText = 'Mengubah password untuk pengguna: ' + username;
            document.getElementById('reset-new-pass').value = '';
            document.getElementById('reset-modal').classList.remove('hidden');
        }

        function closeResetModal() {
            document.getElementById('reset-modal').classList.add('hidden');
        }

        async function submitResetPassword() {
            const newPass = document.getElementById('reset-new-pass').value;
            if (!newPass) return alert('Password baru tidak boleh kosong!');
            try {
                const res = await fetch('/api/admin/user/reset-password', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({user_id: selectedResetUserID, new_password: newPass})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Password pengguna berhasil diperbarui!');
                    closeResetModal();
                } else {
                    alert('Gagal reset password: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        async function deleteUser(id, username) {
            if (!confirm('Apakah Anda yakin ingin menghapus pengguna ' + username + '?')) return;
            try {
                const res = await fetch('/api/admin/user/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({user_id: id})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Pengguna berhasil dihapus!');
                    loadUsers();
                } else {
                    alert('Gagal hapus user: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        // --- TAB 2: PDFS ---
        async function loadPDFs() {
            const tbody = document.getElementById('pdfs-table-body');
            tbody.innerHTML = '<tr><td colspan="4" class="p-6 text-center text-zinc-500">Memuat berkas PDF...</td></tr>';
            try {
                const res = await fetch('/api/admin/pdfs');
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Gagal memuat berkas PDF');

                if (!data.pdfs || data.pdfs.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="4" class="p-6 text-center text-zinc-500">Belum ada berkas PDF di folder dataset.</td></tr>';
                    return;
                }

                let html = '';
                data.pdfs.forEach(p => {
                    let sizeKB = (p.size_bytes / 1024 / 1024).toFixed(2) + ' MB';
                    html += '<tr class="hover:bg-zinc-800/40">' +
                        '<td class="p-4 font-medium text-white flex items-center gap-2">' +
                            '<i class="fa-solid fa-file-pdf text-red-400"></i>' +
                            '<span>' + escapeHtml(p.filename) + '</span>' +
                        '</td>' +
                        '<td class="p-4 font-mono text-zinc-400">' + sizeKB + '</td>' +
                        '<td class="p-4 text-zinc-400 font-mono">' + escapeHtml(p.mod_time) + '</td>' +
                        '<td class="p-4 text-right">' +
                            '<button onclick="deletePDF(\'' + escapeHtml(p.raw_name) + '\')" class="px-2.5 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-[11px] transition">' +
                                '<i class="fa-solid fa-trash text-[10px]"></i> Hapus Berkas' +
                            '</button>' +
                        '</td>' +
                    '</tr>';
                });
                tbody.innerHTML = html;
            } catch (err) {
                tbody.innerHTML = '<tr><td colspan="4" class="p-6 text-center text-red-400">Error: ' + escapeHtml(err.message) + '</td></tr>';
            }
        }

        function openUploadPDFModal() {
            document.getElementById('upload-pdf-modal').classList.remove('hidden');
        }

        function closeUploadPDFModal() {
            document.getElementById('upload-pdf-modal').classList.add('hidden');
        }

        async function submitUploadPDF() {
            const input = document.getElementById('pdf-file-input');
            if (!input.files || input.files.length === 0) return alert('Pilih berkas PDF terlebih dahulu!');

            const formData = new FormData();
            formData.append('pdf_file', input.files[0]);

            try {
                const res = await fetch('/api/admin/pdf/upload', {
                    method: 'POST',
                    body: formData
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Berkas PDF berhasil diupload ke dataset!');
                    closeUploadPDFModal();
                    loadPDFs();
                } else {
                    alert('Gagal upload: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Terjadi kesalahan jaringan.');
            }
        }

        async function deletePDF(rawName) {
            if (!confirm('Apakah Anda yakin ingin menghapus berkas ' + rawName + '?')) return;
            try {
                const res = await fetch('/api/admin/pdf/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({raw_name: rawName})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Berkas PDF berhasil dihapus!');
                    loadPDFs();
                } else {
                    alert('Gagal hapus PDF: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        // --- TAB 3: QUIZ CATEGORIES ---
        let availablePDFsCache = [];

        async function loadQuizzes() {
            const container = document.getElementById('quizzes-container');
            container.innerHTML = '<p class="text-xs text-zinc-500 text-center py-6 col-span-2">Memuat kategori kuis AI...</p>';
            try {
                const res = await fetch('/api/admin/quiz-categories');
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Gagal memuat kategori kuis');

                if (!data.categories || data.categories.length === 0) {
                    container.innerHTML = '<p class="text-xs text-zinc-500 text-center py-6 col-span-2">Belum ada Kategori Kuis AI yang dibuat. Klik "Buat Kategori Kuis Baru" di atas.</p>';
                    return;
                }

                let html = '';
                data.categories.forEach(c => {
                    let booksBadges = (c.selected_books || []).map(b => 
                        '<span class="px-2 py-0.5 text-[10px] font-mono bg-zinc-800 text-zinc-300 border border-zinc-700 rounded-md">' + escapeHtml(b) + '</span>'
                    ).join(' ');

                    html += '<div class="bg-zinc-900 border border-zinc-800 p-5 rounded-2xl space-y-4 shadow-md flex flex-col justify-between">' +
                        '<div class="space-y-2.5">' +
                            '<div class="flex justify-between items-start">' +
                                '<div class="flex items-center gap-2 flex-wrap">' +
                                    '<span class="px-2.5 py-0.5 text-[10px] font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full">' + escapeHtml(c.grade) + '</span>' +
                                    '<span class="px-2.5 py-0.5 text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full">' + c.total_questions + ' Soal AI</span>' +
                                '</div>' +
                                '<button onclick="deleteCategory(' + c.id + ')" class="text-red-400 hover:text-red-300 text-xs p-1">' +
                                    '<i class="fa-solid fa-trash"></i> Hapus' +
                                '</button>' +
                            '</div>' +
                            '<h4 class="text-sm font-bold text-white leading-snug">' + escapeHtml(c.name) + '</h4>' +
                            (c.description ? '<p class="text-xs text-zinc-400">' + escapeHtml(c.description) + '</p>' : '') +
                            '<div class="pt-2 border-t border-zinc-800/80">' +
                                '<label class="block text-[10px] font-semibold text-zinc-500 uppercase tracking-wider mb-1.5">Buku Referensi PDF Terpilih:</label>' +
                                '<div class="flex flex-wrap gap-1">' + (booksBadges || '<span class="text-[10px] text-zinc-500">Semua Buku (Default)</span>') + '</div>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                });
                container.innerHTML = html;
            } catch (err) {
                container.innerHTML = '<p class="text-xs text-red-400 text-center py-6 col-span-2">Error: ' + escapeHtml(err.message) + '</p>';
            }
        }

        async function openCreateCategoryModal() {
            const container = document.getElementById('cat-books-checkboxes');
            container.innerHTML = '<p class="text-xs text-zinc-500">Memuat daftar buku...</p>';
            document.getElementById('create-quiz-modal').classList.remove('hidden');

            try {
                const res = await fetch('/api/admin/pdfs');
                const data = await res.json();
                if (res.ok && data.pdfs) {
                    availablePDFsCache = data.pdfs;
                    let html = '';
                    data.pdfs.forEach(p => {
                        html += '<label class="flex items-center space-x-2 text-zinc-300 cursor-pointer hover:text-white p-1 rounded hover:bg-zinc-900 transition">' +
                            '<input type="checkbox" name="cat-book" value="' + escapeHtml(p.filename) + '" class="rounded bg-zinc-900 border-zinc-700 text-blue-600 focus:ring-0">' +
                            '<span class="truncate max-w-xs">' + escapeHtml(p.filename) + '</span>' +
                        '</label>';
                    });
                    container.innerHTML = html || '<p class="text-xs text-zinc-500">Tidak ada berkas PDF di dataset.</p>';
                }
            } catch(e) {
                container.innerHTML = '<p class="text-xs text-red-400">Gagal memuat daftar buku.</p>';
            }
        }

        function closeCreateQuizModal() {
            document.getElementById('create-quiz-modal').classList.add('hidden');
        }

        async function submitCreateCategory() {
            const name = document.getElementById('new-cat-name').value.trim();
            const grade = document.getElementById('new-cat-grade').value;
            const totalQuestions = parseInt(document.getElementById('new-cat-questions').value) || 5;
            const desc = document.getElementById('new-cat-desc').value.trim();

            const selectedBooks = [];
            document.querySelectorAll('input[name="cat-book"]:checked').forEach(cb => {
                selectedBooks.push(cb.value);
            });

            if (!name) return alert('Nama Kategori Kuis wajib diisi!');

            try {
                const res = await fetch('/api/admin/quiz-category/create', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        name: name,
                        grade: grade,
                        selected_books: selectedBooks,
                        total_questions: totalQuestions,
                        description: desc
                    })
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Kategori Kuis AI berhasil dibuat!');
                    closeCreateQuizModal();
                    loadQuizzes();
                } else {
                    alert('Gagal menambah kategori: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        async function deleteCategory(id) {
            if (!confirm('Apakah Anda yakin ingin menghapus Kategori Kuis ini?')) return;
            try {
                const res = await fetch('/api/admin/quiz-category/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({id: id})
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Kategori Kuis berhasil dihapus!');
                    loadQuizzes();
                } else {
                    alert('Gagal hapus kategori: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal terhubung ke server.');
            }
        }

        // --- TAB 4: CONFIG ---
        document.getElementById('config-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const baseUrl = document.getElementById('cfg-base-url').value;
            const apiKey = document.getElementById('cfg-api-key').value;
            const model = document.getElementById('cfg-model').value;
            const sysPrompt = document.getElementById('cfg-system-prompt').value;
            const mayarKey = document.getElementById('cfg-mayar-key') ? document.getElementById('cfg-mayar-key').value : '';

            try {
                const res = await fetch('/api/admin/config', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        base_url: baseUrl,
                        api_key: apiKey,
                        model: model,
                        system_prompt: sysPrompt,
                        mayar_api_key: mayarKey
                    })
                });
                const data = await res.json();
                if (res.ok && data.success) {
                    alert('Konfigurasi API AI berhasil diperbarui!');
                } else {
                    alert('Gagal update config: ' + (data.error || 'Terjadi kesalahan'));
                }
            } catch(err) {
                alert('Gagal menyimpan konfigurasi.');
            }
        });

        async function fetchVisionModels() {
            const btn = document.getElementById('btn-fetch-models');
            const icon = document.getElementById('fetch-model-icon');
            const select = document.getElementById('cfg-model-select');
            const input = document.getElementById('cfg-model');

            if (btn) btn.disabled = true;
            if (icon) icon.classList.add('animate-spin');

            try {
                const res = await fetch('/api/admin/models');
                const data = await res.json();

                if (res.ok && data.models && data.models.length > 0) {
                    select.innerHTML = '<option value="">-- Pilih Model Vision (' + data.models.length + ') --</option>';
                    data.models.forEach(m => {
                        const opt = document.createElement('option');
                        opt.value = m;
                        opt.innerText = m;
                        if (m === input.value) opt.selected = true;
                        select.appendChild(opt);
                    });
                    alert('Berhasil mengambil ' + data.models.length + ' model AI berdukungan Vision/Multimodal dari Provider API!');
                } else {
                    alert('Gagal fetch model: ' + (data.error || 'Provider tidak mengembalikan daftar model.'));
                }
            } catch(e) {
                alert('Gagal terhubung ke server provider API.');
            } finally {
                if (btn) btn.disabled = false;
                if (icon) icon.classList.remove('animate-spin');
            }
        }
        window.fetchVisionModels = fetchVisionModels;

        // Initialize active tab from server parameter
        switchTab('{{.ActiveTab}}');
    </script>
</body>
</html>`

	currCfg := config.GetConfig()
	t, _ := template.New("admin").Parse(tmpl)
	t.Execute(w, map[string]interface{}{"Config": currCfg, "ActiveTab": activeTab})
}
