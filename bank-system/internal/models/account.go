package models

import (
	"errors"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
)

type AccountType string

const (
	AccountTypeDebit  AccountType = "DEBIT"
	AccountTypeCredit AccountType = "CREDIT"
)

type Account struct {
	ID        int64       `json:"id" db:"id"`
	UserID    int64       `json:"user_id" db:"user_id"`
	Number    string      `json:"number" db:"number"`
	Type      AccountType `json:"type" db:"type"`
	Balance   float64     `json:"balance" db:"balance"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type AccountCreation struct {
	Type AccountType `json:"type"`
}

type AccountResponse struct {
	ID        int64       `json:"id"`
	Number    string      `json:"number"`
	Type      AccountType `json:"type"`
	Balance   float64     `json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
}

type DepositRequest struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type WithdrawRequest struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type BalancePrediction struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
	Events  []string  `json:"events,omitempty"`
}

func (a *Account) CanWithdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if a.Balance < amount {
		return ErrInsufficientFunds
	}

	return nil
}

func ToAccountResponse(account Account) AccountResponse {
	return AccountResponse{
		ID:        account.ID,
		Number:    account.Number,
		Type:      account.Type,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
	}
}
