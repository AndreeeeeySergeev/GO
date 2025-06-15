package service

import (
	"time"

	"bank-service/internal/models"
	"bank-service/internal/repository"
)

type AnalyticsService interface {
	GetTransactionAnalytics(userID int64, period string) (models.TransactionAnalytics, error)
	GetCreditAnalytics(userID int64) (models.CreditAnalytics, error)
}

type analyticsService struct {
	transactionRepo repository.TransactionRepository
	creditRepo      repository.CreditRepository
	paymentRepo     repository.PaymentRepository
}

func NewAnalyticsService(
	transactionRepo repository.TransactionRepository,
	creditRepo repository.CreditRepository,
	paymentRepo repository.PaymentRepository,
) AnalyticsService {
	return &analyticsService{
		transactionRepo: transactionRepo,
		creditRepo:      creditRepo,
		paymentRepo:     paymentRepo,
	}
}

func (s *analyticsService) GetTransactionAnalytics(userID int64, period string) (models.TransactionAnalytics, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	endDate = now

	transactions, err := s.transactionRepo.GetUserTransactionsByPeriod(userID, startDate, endDate)
	if err != nil {
		return models.TransactionAnalytics{}, err
	}

	var totalIncome, totalExpense float64
	categoryBreakdown := make(map[string]float64)

	dailyMap := make(map[string]models.DailyTransactionSummary)

	for _, tx := range transactions {
		date := tx.TransactionDate.Format("2006-01-02")

		daily, exists := dailyMap[date]
		if !exists {
			daily = models.DailyTransactionSummary{
				Date:    tx.TransactionDate,
				Income:  0,
				Expense: 0,
			}
		}

		switch tx.Type {
		case models.TransactionTypeDeposit:
			totalIncome += tx.Amount
			daily.Income += tx.Amount
			categoryBreakdown["Income"] += tx.Amount

		case models.TransactionTypeWithdraw, models.TransactionTypePayment:
			totalExpense += tx.Amount
			daily.Expense += tx.Amount

			if tx.Description == "Withdrawal from account" {
				categoryBreakdown["Cash"] += tx.Amount
			} else if tx.Description == "Card payment" {
				categoryBreakdown["Shopping"] += tx.Amount
			} else {
				categoryBreakdown["Other"] += tx.Amount
			}

		case models.TransactionTypeTransfer:
			if tx.FromAccountID != nil {
				totalExpense += tx.Amount
				daily.Expense += tx.Amount
				categoryBreakdown["Transfers"] += tx.Amount
			}

			if tx.ToAccountID != nil {
				totalIncome += tx.Amount
				daily.Income += tx.Amount
			}
		}

		dailyMap[date] = daily
	}

	var dailyTransactions []models.DailyTransactionSummary
	for _, daily := range dailyMap {
		dailyTransactions = append(dailyTransactions, daily)
	}

	return models.TransactionAnalytics{
		TotalIncome:       totalIncome,
		TotalExpense:      totalExpense,
		CategoryBreakdown: categoryBreakdown,
		DailyTransactions: dailyTransactions,
	}, nil
}

func (s *analyticsService) GetCreditAnalytics(userID int64) (models.CreditAnalytics, error) {
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		return models.CreditAnalytics{}, err
	}

	var totalDebt, monthlyPayments float64
	activeCredits := 0

	for _, credit := range credits {
		if credit.Status == models.CreditStatusActive || credit.Status == models.CreditStatusOverdue {
			activeCredits++

			schedules, err := s.paymentRepo.GetByCreditID(credit.ID)
			if err != nil {
				continue
			}

			for _, schedule := range schedules {
				if schedule.Status == models.PaymentStatusPending {
					totalDebt += schedule.Amount

					now := time.Now()
					if schedule.PaymentDate.Year() == now.Year() && schedule.PaymentDate.Month() == now.Month() {
						monthlyPayments += schedule.Amount
					}
				}
			}
		}
	}

	debtToIncomeRatio := 0.0
	if monthlyPayments > 0 {
		estimatedMonthlyIncome := 100000.0
		debtToIncomeRatio = monthlyPayments / estimatedMonthlyIncome
	}

	return models.CreditAnalytics{
		TotalDebt:         totalDebt,
		MonthlyPayments:   monthlyPayments,
		DebtToIncomeRatio: debtToIncomeRatio,
		RemainingCredits:  activeCredits,
	}, nil
}
