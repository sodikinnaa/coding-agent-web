package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"coding_agent_web/internal/config"
	"coding_agent_web/internal/db"
	"coding_agent_web/internal/model"
)

// Credit Package Management

func GetCreditPackages() ([]model.CreditPackageItem, error) {
	rows, err := db.DB.Query("SELECT id, name, daily_limit, price, description, is_active, created_at FROM credit_packages WHERE is_active = 1 ORDER BY price ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CreditPackageItem
	for rows.Next() {
		var item model.CreditPackageItem
		var activeInt int
		if err := rows.Scan(&item.ID, &item.Name, &item.DailyLimit, &item.Price, &item.Description, &activeInt, &item.CreatedAt); err == nil {
			item.IsActive = (activeInt == 1)
			list = append(list, item)
		}
	}
	return list, nil
}

func GetAllCreditPackagesAdmin() ([]model.CreditPackageItem, error) {
	rows, err := db.DB.Query("SELECT id, name, daily_limit, price, description, is_active, created_at FROM credit_packages ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CreditPackageItem
	for rows.Next() {
		var item model.CreditPackageItem
		var activeInt int
		if err := rows.Scan(&item.ID, &item.Name, &item.DailyLimit, &item.Price, &item.Description, &activeInt, &item.CreatedAt); err == nil {
			item.IsActive = (activeInt == 1)
			list = append(list, item)
		}
	}
	return list, nil
}

func CreateCreditPackageAdmin(item model.CreditPackageItem) error {
	_, err := db.DB.Exec("INSERT INTO credit_packages (name, daily_limit, price, description, is_active) VALUES (?, ?, ?, ?, 1)",
		item.Name, item.DailyLimit, item.Price, item.Description)
	return err
}

func UpdateCreditPackageAdmin(item model.CreditPackageItem) error {
	_, err := db.DB.Exec("UPDATE credit_packages SET name = ?, daily_limit = ?, price = ?, description = ? WHERE id = ?",
		item.Name, item.DailyLimit, item.Price, item.Description, item.ID)
	return err
}

func DeleteCreditPackageAdmin(id int64) error {
	_, err := db.DB.Exec("DELETE FROM credit_packages WHERE id = ?", id)
	return err
}

// Mayar.id Payment Service

func CreateMayarQRISTransaction(userID int64, packageID int64) (*model.PaymentTransaction, error) {
	currCfg := config.GetConfig()
	if strings.TrimSpace(currCfg.MayarAPIKey) == "" {
		return nil, fmt.Errorf("Gateway Pembayaran Mayar.id belum di-konfigurasi oleh Admin! Silakan hubungi Administrator.")
	}

	var pkg model.CreditPackageItem
	var activeInt int
	err := db.DB.QueryRow("SELECT id, name, daily_limit, price, description, is_active FROM credit_packages WHERE id = ?", packageID).
		Scan(&pkg.ID, &pkg.Name, &pkg.DailyLimit, &pkg.Price, &pkg.Description, &activeInt)
	if err != nil {
		return nil, fmt.Errorf("Paket kredit tidak ditemukan")
	}

	var userEmail string
	db.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&userEmail)
	if !strings.Contains(userEmail, "@") {
		userEmail = userEmail + "@kurikulum.app"
	}

	txID := fmt.Sprintf("TX-%d-%d", userID, time.Now().Unix())

	qrURL := ""
	mayarReqBody, _ := json.Marshal(map[string]interface{}{
		"name":   fmt.Sprintf("Paket Kredit Chat AI - %s Tier", pkg.Name),
		"email":  userEmail,
		"amount": pkg.Price,
	})

	req, err := http.NewRequest("POST", "https://api.mayar.id/hl/v1/qrcode/create", bytes.NewBuffer(mayarReqBody))
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+currCfg.MayarAPIKey)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < 300 {
			defer resp.Body.Close()
			var mRes struct {
				Data struct {
					URL  string `json:"url"`
					Link string `json:"link"`
					QR   string `json:"qr"`
				} `json:"data"`
			}
			if json.NewDecoder(resp.Body).Decode(&mRes) == nil {
				if mRes.Data.URL != "" {
					qrURL = mRes.Data.URL
				} else if mRes.Data.QR != "" {
					qrURL = mRes.Data.QR
				} else if mRes.Data.Link != "" {
					qrURL = mRes.Data.Link
				}
			}
		} else if resp != nil {
			bodyB, _ := io.ReadAll(resp.Body)
			log.Printf("Mayar API Error (%d): %s", resp.StatusCode, string(bodyB))
		}
	}

	if qrURL == "" {
		return nil, fmt.Errorf("Gagal menghubungi server Mayar.id (API Key Mayar tidak valid atau terjadi error dari Mayar)")
	}

	expiredAt := time.Now().Add(1 * time.Hour)

	_, err = db.DB.Exec("INSERT INTO payment_transactions (id, user_id, tier_name, daily_limit, amount, status, qr_url, expired_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		txID, userID, pkg.Name, pkg.DailyLimit, pkg.Price, "pending", qrURL, expiredAt)

	if err != nil {
		return nil, err
	}

	return &model.PaymentTransaction{
		ID:         txID,
		UserID:     userID,
		TierName:   pkg.Name,
		DailyLimit: pkg.DailyLimit,
		Amount:     pkg.Price,
		Status:     "pending",
		QRURL:      qrURL,
		CreatedAt:  time.Now(),
		ExpiredAt:  &expiredAt,
	}, nil
}

func ProcessMayarPaymentSuccess(txID string) error {
	var userID int64
	var limit int
	var status string
	err := db.DB.QueryRow("SELECT user_id, daily_limit, status FROM payment_transactions WHERE id = ?", txID).Scan(&userID, &limit, &status)
	if err != nil {
		return err
	}

	if status == "paid" {
		return nil
	}

	_, err = db.DB.Exec("UPDATE users SET daily_limit = ? WHERE id = ?", limit, userID)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec("UPDATE payment_transactions SET status = 'paid', paid_at = CURRENT_TIMESTAMP WHERE id = ?", txID)
	return err
}

func GetPaymentTransactionStatus(txID string) (*model.PaymentTransaction, error) {
	var tx model.PaymentTransaction
	var expStr sql.NullString
	err := db.DB.QueryRow("SELECT id, user_id, tier_name, daily_limit, amount, status, qr_url, created_at, expired_at FROM payment_transactions WHERE id = ?", txID).
		Scan(&tx.ID, &tx.UserID, &tx.TierName, &tx.DailyLimit, &tx.Amount, &tx.Status, &tx.QRURL, &tx.CreatedAt, &expStr)
	if err != nil {
		return nil, err
	}

	if expStr.Valid && expStr.String != "" {
		t, err := time.Parse("2006-01-02 15:04:05", expStr.String)
		if err == nil {
			tx.ExpiredAt = &t
			if tx.Status == "pending" && time.Now().After(t) {
				tx.Status = "expired"
				db.DB.Exec("UPDATE payment_transactions SET status = 'expired' WHERE id = ?", tx.ID)
			}
		}
	}

	return &tx, nil
}

func GetUserPaymentTransactions(userID int64) ([]model.PaymentTransaction, error) {
	rows, err := db.DB.Query("SELECT id, user_id, tier_name, daily_limit, amount, status, qr_url, created_at, expired_at FROM payment_transactions WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.PaymentTransaction
	for rows.Next() {
		var tx model.PaymentTransaction
		var expStr sql.NullString
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.TierName, &tx.DailyLimit, &tx.Amount, &tx.Status, &tx.QRURL, &tx.CreatedAt, &expStr); err == nil {
			if expStr.Valid && expStr.String != "" {
				t, err := time.Parse("2006-01-02 15:04:05", expStr.String)
				if err == nil {
					tx.ExpiredAt = &t
					if tx.Status == "pending" && time.Now().After(t) {
						tx.Status = "expired"
						db.DB.Exec("UPDATE payment_transactions SET status = 'expired' WHERE id = ?", tx.ID)
					}
				}
			}
			list = append(list, tx)
		}
	}
	return list, nil
}


func CancelPaymentTransactionUser(userID int64, txID string) error {
	_, err := db.DB.Exec("UPDATE payment_transactions SET status = 'cancelled' WHERE id = ? AND user_id = ? AND status = 'pending'", txID, userID)
	return err
}

func DeletePaymentTransactionUser(userID int64, txID string) error {
	_, err := db.DB.Exec("DELETE FROM payment_transactions WHERE id = ? AND user_id = ?", txID, userID)
	return err
}
