package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"

	"coding_agent_web/internal/model"
)

const ConfigPath = "/root/coding_agent_web/config.json"
const EncPrefix = "ENC:AES256GCM:"
const SecretMasterKey = "antigravity-ai-kurikulum-master-secret-key-v1"

var (
	mutex sync.Mutex
	cfg   model.AppConfig
)

// Derive 32-byte (256-bit) AES key from master secret phrase
func getAESKey() []byte {
	hash := sha256.Sum256([]byte(SecretMasterKey))
	return hash[:]
}

// EncryptToken encrypts raw secret tokens using AES-256-GCM
func EncryptToken(plainText string) string {
	plainText = strings.TrimSpace(plainText)
	if plainText == "" || strings.HasPrefix(plainText, EncPrefix) {
		return plainText
	}

	key := getAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return plainText
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plainText
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plainText
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return EncPrefix + base64.StdEncoding.EncodeToString(cipherText)
}

// DecryptToken decrypts cipherText encrypted with EncryptToken
func DecryptToken(cipherText string) string {
	cipherText = strings.TrimSpace(cipherText)
	if !strings.HasPrefix(cipherText, EncPrefix) {
		return cipherText
	}

	rawB64 := strings.TrimPrefix(cipherText, EncPrefix)
	data, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return cipherText
	}

	key := getAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return cipherText
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return cipherText
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return cipherText
	}

	nonce, actualCipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return cipherText
	}

	return string(plainText)
}

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
			var fileCfg model.AppConfig
			if json.Unmarshal(data, &fileCfg) == nil {
				// Decrypt AES-256-GCM tokens from disk
				fileCfg.APIKey = DecryptToken(fileCfg.APIKey)
				fileCfg.MayarAPIKey = DecryptToken(fileCfg.MayarAPIKey)
				cfg = fileCfg
			}
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
	// Ensure memory config holds decrypted token values
	cfg = newCfg
	cfg.APIKey = DecryptToken(newCfg.APIKey)
	cfg.MayarAPIKey = DecryptToken(newCfg.MayarAPIKey)

	saveConfigNoLock()
}

func saveConfigNoLock() {
	// Encrypt secret tokens before saving to JSON file
	diskCfg := cfg
	diskCfg.APIKey = EncryptToken(cfg.APIKey)
	diskCfg.MayarAPIKey = EncryptToken(cfg.MayarAPIKey)

	data, _ := json.MarshalIndent(diskCfg, "", "  ")
	_ = os.WriteFile(ConfigPath, data, 0600)
}
