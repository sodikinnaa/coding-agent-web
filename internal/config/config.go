package config

import (
	"encoding/json"
	"os"
	"sync"
	"coding_agent_web/internal/model"
)

const ConfigPath = "/root/coding_agent_web/config.json"

var (
	mutex sync.Mutex
	cfg   model.AppConfig
)

func LoadConfig() model.AppConfig {
	mutex.Lock()
	defer mutex.Unlock()

	cfg = model.AppConfig{
		AdminPassword: "L4njutk4n",
		BaseURL:       "https://member.wakdondin.my.id/api/v1",
		APIKey:        "wak_i9xNSTqKCUyPp8YHgFO3T94RxtmQ2IlAn2Qq7ZZMQ3KigowPQfchDcNlY8zJtdGj",
		Model:         "gemini-3.6-flash-high",
		SystemPrompt:  "Kamu adalah Asisten AI Kurikulum Koding & AI Indonesia (SD-SMA). PENTING: Jawab pertanyaan pengguna berdasarkan dokumen Kurikulum Koding & AI yang dilampirkan. Selalu cantumkan referensi nama buku dan nomor halaman persis apabila terdapat pada materi!",
	}

	if _, err := os.Stat(ConfigPath); err == nil {
		data, err := os.ReadFile(ConfigPath)
		if err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
	} else {
		saveConfigNoLock()
	}
	return cfg
}

func GetConfig() model.AppConfig {
	mutex.Lock()
	defer mutex.Unlock()
	return cfg
}

func SaveConfig(newCfg model.AppConfig) {
	mutex.Lock()
	defer mutex.Unlock()
	cfg = newCfg
	saveConfigNoLock()
}

func saveConfigNoLock() {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(ConfigPath, data, 0644)
}
