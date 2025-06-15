package models

import (
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusOverdue  PaymentStatus = "OVERDUE"
	PaymentStatusCanceled PaymentStatus = "CANCELED"
)

type PaymentSchedule struct {
	ID            int64         `json:"id" db:"id"`
	CreditID      int64         `json:"credit_id" db:"credit_id"`
	PaymentDate   time.Time     `json:"payment_date" db:"payment_date"`
	Amount        float64       `json:"amount" db:"amount"`
	Principal     float64       `json:"principal" db:"principal"`
	Interest      float64       `json:"interest" db:"interest"`
	RemainingDebt float64       `json:"remaining_debt" db:"remaining_debt"`
	Status        PaymentStatus `json:"status" db:"status"`
	PaidDate      *time.Time    `json:"paid_date,omitempty" db:"paid_date"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type PaymentScheduleResponse struct {
	PaymentDate   time.Time     `json:"payment_date"`
	Amount        float64       `json:"amount"`
	Principal     float64       `json:"principal"`
	Interest      float64       `json:"interest"`
	RemainingDebt float64       `json:"remaining_debt"`
	Status        PaymentStatus `json:"status"`
	PaidDate      *time.Time    `json:"paid_date,omitempty"`
}

func ToPaymentScheduleResponse(schedule PaymentSchedule) PaymentScheduleResponse {
	return PaymentScheduleResponse{
		PaymentDate:   schedule.PaymentDate,
		Amount:        schedule.Amount,
		Principal:     schedule.Principal,
		Interest:      schedule.Interest,
		RemainingDebt: schedule.RemainingDebt,
		Status:        schedule.Status,
		PaidDate:      schedule.PaidDate,
	}
}
