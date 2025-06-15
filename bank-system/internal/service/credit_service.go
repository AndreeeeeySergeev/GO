package service

import (
	"errors"
	"math"
	"time"

	"bank-service/internal/models"
	"bank-service/internal/repository"
)

var (
	ErrCreditNotFound      = errors.New("credit not found")
	ErrCreditAccessDenied  = errors.New("access to this credit is denied")
	ErrInvalidCreditAmount = errors.New("credit amount must be positive")
	ErrInvalidCreditTerm   = errors.New("credit term must be between 3 and 60 months")
)

type CreditService interface {
	Apply(userID int64, application models.CreditApplication) (models.CreditResponse, error)
	GetByID(id int64, userID int64) (models.CreditResponse, error)
	GetByUserID(userID int64) ([]models.CreditResponse, error)
	GetSchedule(creditID int64, userID int64) ([]models.PaymentScheduleResponse, error)
	ProcessPendingPayments() error
}

type creditService struct {
	creditRepo   repository.CreditRepository
	paymentRepo  repository.PaymentRepository
	accountRepo  repository.AccountRepository
	cbrService   CBRService
	emailService EmailService
}

func NewCreditService(
	creditRepo repository.CreditRepository,
	paymentRepo repository.PaymentRepository,
	accountRepo repository.AccountRepository,
	cbrService CBRService,
	emailService EmailService,
) CreditService {
	return &creditService{
		creditRepo:   creditRepo,
		paymentRepo:  paymentRepo,
		accountRepo:  accountRepo,
		cbrService:   cbrService,
		emailService: emailService,
	}
}

func (s *creditService) Apply(userID int64, application models.CreditApplication) (models.CreditResponse, error) {
	if application.Amount <= 0 {
		return models.CreditResponse{}, ErrInvalidCreditAmount
	}

	if application.Term < 3 || application.Term > 60 {
		return models.CreditResponse{}, ErrInvalidCreditTerm
	}

	account, err := s.accountRepo.GetByID(application.AccountID)
	if err != nil {
		return models.CreditResponse{}, ErrAccountNotFound
	}

	if account.UserID != userID {
		return models.CreditResponse{}, ErrAccountAccessDenied
	}

	// Получение ключевой ставки ЦБ РФ
	keyRate, err := s.cbrService.GetKeyRate()
	if err != nil {
		keyRate = 7.5
	}

	interestRate := keyRate + 5.0

	monthlyInterestRate := interestRate / 100 / 12
	monthlyPayment := application.Amount * monthlyInterestRate * math.Pow(1+monthlyInterestRate, float64(application.Term)) /
		(math.Pow(1+monthlyInterestRate, float64(application.Term)) - 1)

	totalPayment := monthlyPayment * float64(application.Term)

	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, application.Term, 0)

	credit := models.Credit{
		UserID:         userID,
		AccountID:      application.AccountID,
		Amount:         application.Amount,
		Term:           application.Term,
		InterestRate:   interestRate,
		MonthlyPayment: monthlyPayment,
		TotalPayment:   totalPayment,
		Status:         models.CreditStatusApproved,
		StartDate:      startDate,
		EndDate:        endDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	tx, err := s.creditRepo.BeginTx()
	if err != nil {
		return models.CreditResponse{}, err
	}
	defer tx.Rollback()

	creditID, err := s.creditRepo.CreateTx(tx, credit)
	if err != nil {
		return models.CreditResponse{}, err
	}

	credit.ID = creditID

	paymentSchedules, err := s.generatePaymentSchedule(credit)
	if err != nil {
		return models.CreditResponse{}, err
	}

	if err := s.paymentRepo.CreateBatchTx(tx, paymentSchedules); err != nil {
		return models.CreditResponse{}, err
	}

	newBalance := account.Balance + application.Amount
	if err := s.accountRepo.UpdateBalanceTx(tx, account.ID, newBalance); err != nil {
		return models.CreditResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.CreditResponse{}, err
	}

	go s.emailService.SendCreditApprovalEmail(
		credit.UserID,
		credit.Amount,
		credit.InterestRate,
		credit.MonthlyPayment,
		credit.Term,
	)

	return models.ToCreditResponse(credit), nil
}

func (s *creditService) GetByID(id int64, userID int64) (models.CreditResponse, error) {
	credit, err := s.creditRepo.GetByID(id)
	if err != nil {
		return models.CreditResponse{}, ErrCreditNotFound
	}

	if credit.UserID != userID {
		return models.CreditResponse{}, ErrCreditAccessDenied
	}

	return models.ToCreditResponse(credit), nil
}

func (s *creditService) GetByUserID(userID int64) ([]models.CreditResponse, error) {
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []models.CreditResponse
	for _, credit := range credits {
		response = append(response, models.ToCreditResponse(credit))
	}

	return response, nil
}

func (s *creditService) GetSchedule(creditID int64, userID int64) ([]models.PaymentScheduleResponse, error) {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		return nil, ErrCreditNotFound
	}

	if credit.UserID != userID {
		return nil, ErrCreditAccessDenied
	}

	schedules, err := s.paymentRepo.GetByCreditID(creditID)
	if err != nil {
		return nil, err
	}

	var response []models.PaymentScheduleResponse
	for _, schedule := range schedules {
		response = append(response, models.ToPaymentScheduleResponse(schedule))
	}

	return response, nil
}

func (s *creditService) ProcessPendingPayments() error {
	pendingPayments, err := s.paymentRepo.GetPendingPayments()
	if err != nil {
		return err
	}

	for _, payment := range pendingPayments {
		credit, err := s.creditRepo.GetByID(payment.CreditID)
		if err != nil {
			continue
		}

		account, err := s.accountRepo.GetByID(credit.AccountID)
		if err != nil {
			continue
		}

		tx, err := s.accountRepo.BeginTx()
		if err != nil {
			continue
		}

		if account.Balance >= payment.Amount {
			newBalance := account.Balance - payment.Amount
			if err := s.accountRepo.UpdateBalanceTx(tx, account.ID, newBalance); err != nil {
				tx.Rollback()
				continue
			}

			now := time.Now()
			payment.Status = models.PaymentStatusPaid
			payment.PaidDate = &now

			if err := tx.Commit(); err != nil {
				continue
			}

			s.paymentRepo.UpdateStatus(payment.ID, models.PaymentStatusPaid, &now)

			go s.emailService.SendPaymentSuccessEmail(credit.UserID, payment.Amount, credit.ID)
		} else {
			tx.Rollback()

			payment.Status = models.PaymentStatusOverdue
			s.paymentRepo.UpdateStatus(payment.ID, models.PaymentStatusOverdue, nil)

			s.creditRepo.UpdateStatus(credit.ID, models.CreditStatusOverdue)

			go s.emailService.SendPaymentOverdueEmail(credit.UserID, payment.Amount, credit.ID)
		}
	}

	return nil
}

func (s *creditService) generatePaymentSchedule(credit models.Credit) ([]models.PaymentSchedule, error) {
	var schedules []models.PaymentSchedule

	remainingDebt := credit.Amount
	monthlyInterestRate := credit.InterestRate / 100 / 12

	for i := 0; i < credit.Term; i++ {
		paymentDate := credit.StartDate.AddDate(0, i+1, 0)

		interestPayment := remainingDebt * monthlyInterestRate

		principalPayment := credit.MonthlyPayment - interestPayment

		remainingDebt -= principalPayment

		if i == credit.Term-1 {
			principalPayment += remainingDebt
			remainingDebt = 0
		}

		now := time.Now()
		schedule := models.PaymentSchedule{
			CreditID:      credit.ID,
			PaymentDate:   paymentDate,
			Amount:        credit.MonthlyPayment,
			Principal:     principalPayment,
			Interest:      interestPayment,
			RemainingDebt: remainingDebt,
			Status:        models.PaymentStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}
