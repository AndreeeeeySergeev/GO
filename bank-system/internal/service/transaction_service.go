package service

import (
	"bank-service/internal/models"
	"bank-service/internal/repository"
)

type TransactionService interface {
	GetByID(id int64, userID int64) (models.TransactionResponse, error)
	GetByUserID(userID int64, limit, offset int) ([]models.TransactionResponse, error)
	GetByAccountID(accountID int64, userID int64, limit, offset int) ([]models.TransactionResponse, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository, accountRepo repository.AccountRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

func (s *transactionService) GetByID(id int64, userID int64) (models.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(id)
	if err != nil {
		return models.TransactionResponse{}, err
	}

	if transaction.UserID != userID {
		return models.TransactionResponse{}, ErrAccountAccessDenied
	}

	return models.ToTransactionResponse(transaction), nil
}

func (s *transactionService) GetByUserID(userID int64, limit, offset int) ([]models.TransactionResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	transactions, err := s.transactionRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var response []models.TransactionResponse
	for _, transaction := range transactions {
		response = append(response, models.ToTransactionResponse(transaction))
	}

	return response, nil
}

func (s *transactionService) GetByAccountID(accountID int64, userID int64, limit, offset int) ([]models.TransactionResponse, error) {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	if account.UserID != userID {
		return nil, ErrAccountAccessDenied
	}

	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	transactions, err := s.transactionRepo.GetByAccountID(accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	var response []models.TransactionResponse
	for _, transaction := range transactions {
		response = append(response, models.ToTransactionResponse(transaction))
	}

	return response, nil
}
