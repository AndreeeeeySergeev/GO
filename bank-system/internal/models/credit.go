package models

import (
	"time"
)

type CreditStatus string

const (
	CreditStatusActive   CreditStatus = "ACTIVE"
	CreditStatusClosed   CreditStatus = "CLOSED"
	CreditStatusOverdue  CreditStatus = "OVERDUE"
	CreditStatusApproved CreditStatus = "APPROVED"
	CreditStatusRejected CreditStatus = "REJECTED"
	CreditStatusPending  CreditStatus = "PENDING"
)

type Credit struct {
	ID             int64        `json:"id" db:"id"`
	UserID         int64        `json:"user_id" db:"user_id"`
	AccountID      int64        `json:"account_id" db:"account_id"`
	Amount         float64      `json:"amount" db:"amount"`
	Term           int          `json:"term" db:"term"`
	InterestRate   float64      `json:"interest_rate" db:"interest_rate"`
	MonthlyPayment float64      `json:"monthly_payment" db:"monthly_payment"`
	TotalPayment   float64      `json:"total_payment" db:"total_payment"`
	Status         CreditStatus `json:"status" db:"status"`
	StartDate      time.Time    `json:"start_date" db:"start_date"`
	EndDate        time.Time    `json:"end_date" db:"end_date"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

type CreditApplication struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
	Term      int     `json:"term"`
}

type CreditResponse struct {
	ID             int64        `json:"id"`
	Amount         float64      `json:"amount"`
	Term           int          `json:"term"`
	InterestRate   float64      `json:"interest_rate"`
	MonthlyPayment float64      `json:"monthly_payment"`
	TotalPayment   float64      `json:"total_payment"`
	Status         CreditStatus `json:"status"`
	StartDate      time.Time    `json:"start_date"`
	EndDate        time.Time    `json:"end_date"`
}

type CreditAnalytics struct {
	TotalDebt         float64 `json:"total_debt"`
	MonthlyPayments   float64 `json:"monthly_payments"`
	DebtToIncomeRatio float64 `json:"debt_to_income_ratio"`
	RemainingCredits  int     `json:"remaining_credits"`
}

func ToCreditResponse(credit Credit) CreditResponse {
	return CreditResponse{
		ID:             credit.ID,
		Amount:         credit.Amount,
		Term:           credit.Term,
		InterestRate:   credit.InterestRate,
		MonthlyPayment: credit.MonthlyPayment,
		TotalPayment:   credit.TotalPayment,
		Status:         credit.Status,
		StartDate:      credit.StartDate,
		EndDate:        credit.EndDate,
	}
}
