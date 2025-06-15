package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"bank-service/internal/models"
	"bank-service/internal/repository"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrAccountAccessDenied = errors.New("access to this account is denied")
	ErrSameAccount         = errors.New("cannot transfer to the same account")
)

type AccountService interface {
	Create(userID int64, accountType models.AccountType) (models.AccountResponse, error)
	GetByID(id int64, userID int64) (models.AccountResponse, error)
	GetByUserID(userID int64) ([]models.AccountResponse, error)
	Deposit(request models.DepositRequest, userID int64) error
	Withdraw(request models.WithdrawRequest, userID int64) error
	Transfer(request models.TransferRequest, userID int64) error
	PredictBalance(accountID int64, userID int64, days int) ([]models.BalancePrediction, error)
}

type accountService struct {
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
}

func NewAccountService(accountRepo repository.AccountRepository, transactionRepo repository.TransactionRepository) AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *accountService) Create(userID int64, accountType models.AccountType) (models.AccountResponse, error) {
	accountNumber := generateAccountNumber()

	now := time.Now()
	account := models.Account{
		UserID:    userID,
		Number:    accountNumber,
		Type:      accountType,
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := s.accountRepo.Create(account)
	if err != nil {
		return models.AccountResponse{}, err
	}

	account.ID = id

	return models.ToAccountResponse(account), nil
}

func (s *accountService) GetByID(id int64, userID int64) (models.AccountResponse, error) {
	account, err := s.accountRepo.GetByID(id)
	if err != nil {
		return models.AccountResponse{}, ErrAccountNotFound
	}

	if account.UserID != userID {
		return models.AccountResponse{}, ErrAccountAccessDenied
	}

	return models.ToAccountResponse(account), nil
}

func (s *accountService) GetByUserID(userID int64) ([]models.AccountResponse, error) {
	accounts, err := s.accountRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []models.AccountResponse
	for _, account := range accounts {
		response = append(response, models.ToAccountResponse(account))
	}

	return response, nil
}

func (s *accountService) Deposit(request models.DepositRequest, userID int64) error {
	if request.Amount <= 0 {
		return ErrInvalidAmount
	}

	account, err := s.accountRepo.GetByID(request.AccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if account.UserID != userID {
		return ErrAccountAccessDenied
	}

	tx, err := s.accountRepo.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	newBalance := account.Balance + request.Amount
	if err := s.accountRepo.UpdateBalanceTx(tx, account.ID, newBalance); err != nil {
		return err
	}

	transaction := models.Transaction{
		UserID:          userID,
		ToAccountID:     &account.ID,
		Type:            models.TransactionTypeDeposit,
		Amount:          request.Amount,
		Description:     "Deposit to account",
		Status:          "COMPLETED",
		TransactionDate: time.Now(),
		CreatedAt:       time.Now(),
	}

	if _, err := s.transactionRepo.CreateTx(tx, transaction); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) Withdraw(request models.WithdrawRequest, userID int64) error {
	if request.Amount <= 0 {
		return ErrInvalidAmount
	}

	account, err := s.accountRepo.GetByID(request.AccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if account.UserID != userID {
		return ErrAccountAccessDenied
	}

	if err := account.CanWithdraw(request.Amount); err != nil {
		return err
	}

	tx, err := s.accountRepo.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	newBalance := account.Balance - request.Amount
	if err := s.accountRepo.UpdateBalanceTx(tx, account.ID, newBalance); err != nil {
		return err
	}

	transaction := models.Transaction{
		UserID:          userID,
		FromAccountID:   &account.ID,
		Type:            models.TransactionTypeWithdraw,
		Amount:          request.Amount,
		Description:     "Withdrawal from account",
		Status:          "COMPLETED",
		TransactionDate: time.Now(),
		CreatedAt:       time.Now(),
	}

	if _, err := s.transactionRepo.CreateTx(tx, transaction); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) Transfer(request models.TransferRequest, userID int64) error {
	if request.Amount <= 0 {
		return ErrInvalidAmount
	}

	if request.FromAccountID == request.ToAccountID {
		return ErrSameAccount
	}

	fromAccount, err := s.accountRepo.GetByID(request.FromAccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if fromAccount.UserID != userID {
		return ErrAccountAccessDenied
	}

	toAccount, err := s.accountRepo.GetByID(request.ToAccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if err := fromAccount.CanWithdraw(request.Amount); err != nil {
		return err
	}

	tx, err := s.accountRepo.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	newFromBalance := fromAccount.Balance - request.Amount
	if err := s.accountRepo.UpdateBalanceTx(tx, fromAccount.ID, newFromBalance); err != nil {
		return err
	}

	newToBalance := toAccount.Balance + request.Amount
	if err := s.accountRepo.UpdateBalanceTx(tx, toAccount.ID, newToBalance); err != nil {
		return err
	}

	transaction := models.Transaction{
		UserID:          userID,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     &toAccount.ID,
		Type:            models.TransactionTypeTransfer,
		Amount:          request.Amount,
		Description:     fmt.Sprintf("Transfer from account %s to account %s", fromAccount.Number, toAccount.Number),
		Status:          "COMPLETED",
		TransactionDate: time.Now(),
		CreatedAt:       time.Now(),
	}

	if _, err := s.transactionRepo.CreateTx(tx, transaction); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) PredictBalance(accountID int64, userID int64, days int) ([]models.BalancePrediction, error) {
	if days <= 0 || days > 365 {
		days = 30 // Значение по умолчанию
	}

	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	if account.UserID != userID {
		return nil, ErrAccountAccessDenied
	}

	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()
	transactions, err := s.transactionRepo.GetUserTransactionsByPeriod(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalIncome, totalExpense float64
	for _, tx := range transactions {
		if tx.Type == models.TransactionTypeDeposit {
			totalIncome += tx.Amount
		} else if tx.Type == models.TransactionTypeWithdraw || tx.Type == models.TransactionTypePayment {
			totalExpense += tx.Amount
		}
	}

	daysInPeriod := 30.0
	avgDailyIncome := totalIncome / daysInPeriod
	avgDailyExpense := totalExpense / daysInPeriod

	predictions := make([]models.BalancePrediction, days)
	currentBalance := account.Balance

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, i+1)

		dailyChange := avgDailyIncome - avgDailyExpense
		currentBalance += dailyChange

		predictions[i] = models.BalancePrediction{
			Date:    date,
			Balance: currentBalance,
			Events:  []string{},
		}
	}

	return predictions, nil
}

func generateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	accountNumber := "4000"
	for i := 0; i < 12; i++ {
		accountNumber += fmt.Sprintf("%d", rand.Intn(10))
	}
	return accountNumber
}
