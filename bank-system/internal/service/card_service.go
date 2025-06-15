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
	ErrCardNotFound     = errors.New("card not found")
	ErrCardAccessDenied = errors.New("access to this card is denied")
	ErrCardInactive     = errors.New("card is inactive")
)

type CardService interface {
	Create(userID int64, request models.CardCreation) (models.CardResponse, error)
	GetByID(id int64, userID int64) (models.CardResponse, error)
	GetByUserID(userID int64) ([]models.CardResponse, error)
	UpdateStatus(id int64, isActive bool, userID int64) error
	ProcessPayment(request models.CardPaymentRequest, userID int64) error
}

type cardService struct {
	cardRepo    repository.CardRepository
	accountRepo repository.AccountRepository
	encryption  EncryptionService
}

func NewCardService(cardRepo repository.CardRepository, accountRepo repository.AccountRepository, encryption EncryptionService) CardService {
	return &cardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		encryption:  encryption,
	}
}

func (s *cardService) Create(userID int64, request models.CardCreation) (models.CardResponse, error) {
	account, err := s.accountRepo.GetByID(request.AccountID)
	if err != nil {
		return models.CardResponse{}, ErrAccountNotFound
	}

	if account.UserID != userID {
		return models.CardResponse{}, ErrAccountAccessDenied
	}

	cardNumber := generateCardNumber()
	expiryDate := generateExpiryDate()
	cvv := generateCVV()

	encryptedNumber, err := s.encryption.EncryptData(cardNumber)
	if err != nil {
		return models.CardResponse{}, err
	}

	encryptedExpiry, err := s.encryption.EncryptData(expiryDate)
	if err != nil {
		return models.CardResponse{}, err
	}

	numberHMAC, err := s.encryption.CreateHMAC(cardNumber)
	if err != nil {
		return models.CardResponse{}, err
	}

	expiryHMAC, err := s.encryption.CreateHMAC(expiryDate)
	if err != nil {
		return models.CardResponse{}, err
	}

	cvvHash, err := s.encryption.HashPassword(cvv)
	if err != nil {
		return models.CardResponse{}, err
	}

	now := time.Now()
	card := models.Card{
		AccountID:  request.AccountID,
		UserID:     userID,
		Number:     encryptedNumber,
		NumberHMAC: numberHMAC,
		ExpiryDate: encryptedExpiry,
		ExpiryHMAC: expiryHMAC,
		CVV:        cvvHash,
		Type:       request.Type,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	id, err := s.cardRepo.Create(card)
	if err != nil {
		return models.CardResponse{}, err
	}

	return models.CardResponse{
		ID:         id,
		AccountID:  request.AccountID,
		Number:     cardNumber,
		ExpiryDate: expiryDate,
		Type:       request.Type,
		IsActive:   true,
		CreatedAt:  now,
	}, nil
}

func (s *cardService) GetByID(id int64, userID int64) (models.CardResponse, error) {
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return models.CardResponse{}, ErrCardNotFound
	}

	if card.UserID != userID {
		return models.CardResponse{}, ErrCardAccessDenied
	}

	decryptedNumber, err := s.encryption.DecryptData(card.Number)
	if err != nil {
		return models.CardResponse{}, err
	}

	if err := s.encryption.VerifyHMAC(decryptedNumber, card.NumberHMAC); err != nil {
		return models.CardResponse{}, errors.New("card data integrity check failed")
	}

	decryptedExpiry, err := s.encryption.DecryptData(card.ExpiryDate)
	if err != nil {
		return models.CardResponse{}, err
	}

	if err := s.encryption.VerifyHMAC(decryptedExpiry, card.ExpiryHMAC); err != nil {
		return models.CardResponse{}, errors.New("card expiry integrity check failed")
	}

	return models.CardResponse{
		ID:         card.ID,
		AccountID:  card.AccountID,
		Number:     decryptedNumber,
		ExpiryDate: decryptedExpiry,
		Type:       card.Type,
		IsActive:   card.IsActive,
		CreatedAt:  card.CreatedAt,
	}, nil
}

func (s *cardService) GetByUserID(userID int64) ([]models.CardResponse, error) {
	cards, err := s.cardRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []models.CardResponse
	for _, card := range cards {
		decryptedNumber, err := s.encryption.DecryptData(card.Number)
		if err != nil {
			continue
		}

		if err := s.encryption.VerifyHMAC(decryptedNumber, card.NumberHMAC); err != nil {
			continue
		}

		decryptedExpiry, err := s.encryption.DecryptData(card.ExpiryDate)
		if err != nil {
			continue
		}

		if err := s.encryption.VerifyHMAC(decryptedExpiry, card.ExpiryHMAC); err != nil {
			continue
		}

		maskedNumber := models.MaskCardNumber(decryptedNumber)

		response = append(response, models.CardResponse{
			ID:         card.ID,
			AccountID:  card.AccountID,
			Number:     maskedNumber,
			ExpiryDate: decryptedExpiry,
			Type:       card.Type,
			IsActive:   card.IsActive,
			CreatedAt:  card.CreatedAt,
		})
	}

	return response, nil
}

func (s *cardService) UpdateStatus(id int64, isActive bool, userID int64) error {
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return ErrCardNotFound
	}

	if card.UserID != userID {
		return ErrCardAccessDenied
	}

	return s.cardRepo.UpdateStatus(id, isActive)
}

func (s *cardService) ProcessPayment(request models.CardPaymentRequest, userID int64) error {
	card, err := s.cardRepo.GetByID(request.CardID)
	if err != nil {
		return ErrCardNotFound
	}

	if card.UserID != userID {
		return ErrCardAccessDenied
	}

	if !card.IsActive {
		return ErrCardInactive
	}

	account, err := s.accountRepo.GetByID(card.AccountID)
	if err != nil {
		return ErrAccountNotFound
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

	return tx.Commit()
}

func generateCardNumber() string {
	rand.Seed(time.Now().UnixNano())

	prefix := []string{"4", "5"}[rand.Intn(2)]

	cardNumber := prefix
	for i := 0; i < 15; i++ {
		cardNumber += fmt.Sprintf("%d", rand.Intn(10))
	}

	// Проверка по алгоритму Луна
	sum := 0
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')
		if (len(cardNumber)-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	if sum%10 != 0 {
		lastDigit := (10 - (sum % 10)) % 10
		cardNumber = cardNumber[:len(cardNumber)-1] + fmt.Sprintf("%d", lastDigit)
	}

	return cardNumber
}

func generateExpiryDate() string {
	now := time.Now()
	month := now.Month()
	year := now.Year() + 3

	return fmt.Sprintf("%02d/%02d", month, year%100)
}

func generateCVV() string {
	rand.Seed(time.Now().UnixNano())

	// Генерация 3-значного CVV
	return fmt.Sprintf("%03d", rand.Intn(1000))
}
