package models

import (
	"time"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "DEPOSIT"
	TransactionTypeWithdraw TransactionType = "WITHDRAW"
	TransactionTypeTransfer TransactionType = "TRANSFER"
	TransactionTypePayment  TransactionType = "PAYMENT"
	TransactionTypeCredit   TransactionType = "CREDIT"
)

type Transaction struct {
	ID              int64           `json:"id" db:"id"`
	UserID          int64           `json:"user_id" db:"user_id"`
	FromAccountID   *int64          `json:"from_account_id,omitempty" db:"from_account_id"`
	ToAccountID     *int64          `json:"to_account_id,omitempty" db:"to_account_id"`
	Type            TransactionType `json:"type" db:"type"`
	Amount          float64         `json:"amount" db:"amount"`
	Description     string          `json:"description" db:"description"`
	Status          string          `json:"status" db:"status"`
	TransactionDate time.Time       `json:"transaction_date" db:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type TransactionResponse struct {
	ID              int64           `json:"id"`
	Type            TransactionType `json:"type"`
	Amount          float64         `json:"amount"`
	Description     string          `json:"description"`
	Status          string          `json:"status"`
	TransactionDate time.Time       `json:"transaction_date"`
}

type TransactionAnalytics struct {
	TotalIncome       float64                   `json:"total_income"`
	TotalExpense      float64                   `json:"total_expense"`
	CategoryBreakdown map[string]float64        `json:"category_breakdown,omitempty"`
	DailyTransactions []DailyTransactionSummary `json:"daily_transactions,omitempty"`
}

type DailyTransactionSummary struct {
	Date    time.Time `json:"date"`
	Income  float64   `json:"income"`
	Expense float64   `json:"expense"`
}

func ToTransactionResponse(transaction Transaction) TransactionResponse {
	return TransactionResponse{
		ID:              transaction.ID,
		Type:            transaction.Type,
		Amount:          transaction.Amount,
		Description:     transaction.Description,
		Status:          transaction.Status,
		TransactionDate: transaction.TransactionDate,
	}
}
