package models

import (
	"time"
)

type CardType string

const (
	CardTypeVirtual  CardType = "VIRTUAL"
	CardTypePhysical CardType = "PHYSICAL"
)

type Card struct {
	ID         int64     `json:"id" db:"id"`
	AccountID  int64     `json:"account_id" db:"account_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Number     string    `json:"-" db:"number_encrypted"`
	NumberHMAC string    `json:"-" db:"number_hmac"`
	ExpiryDate string    `json:"-" db:"expiry_date_encrypted"`
	ExpiryHMAC string    `json:"-" db:"expiry_date_hmac"`
	CVV        string    `json:"-" db:"cvv_hash"`
	Type       CardType  `json:"type" db:"type"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CardCreation struct {
	AccountID int64    `json:"account_id"`
	Type      CardType `json:"type"`
}

type CardResponse struct {
	ID         int64     `json:"id"`
	AccountID  int64     `json:"account_id"`
	Number     string    `json:"number"`
	ExpiryDate string    `json:"expiry_date"`
	Type       CardType  `json:"type"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type CardPaymentRequest struct {
	CardID int64   `json:"card_id"`
	Amount float64 `json:"amount"`
}

// Для безопасного отображения номера карты (только последние 4 цифры)
func MaskCardNumber(number string) string {
	if len(number) < 4 {
		return "****"
	}
	return "****" + number[len(number)-4:]
}
