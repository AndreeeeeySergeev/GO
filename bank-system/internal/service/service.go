package service

import (
	"bank-service/internal/config"
	"bank-service/internal/repository"
)

type Services struct {
	User        UserService
	Account     AccountService
	Card        CardService
	Transaction TransactionService
	Credit      CreditService
	Analytics   AnalyticsService
	CBR         CBRService
	Email       EmailService
	Encryption  EncryptionService
}

type Dependencies struct {
	Repos             *repository.Repositories
	EncryptionService EncryptionService
	EmailService      EmailService
	CBRService        CBRService
	Config            *config.Config
}

func NewServices(deps Dependencies) *Services {
	userService := NewUserService(deps.Repos.User, deps.EncryptionService)
	accountService := NewAccountService(deps.Repos.Account, deps.Repos.Transaction)
	cardService := NewCardService(deps.Repos.Card, deps.Repos.Account, deps.EncryptionService)
	transactionService := NewTransactionService(deps.Repos.Transaction, deps.Repos.Account)
	creditService := NewCreditService(deps.Repos.Credit, deps.Repos.Payment, deps.Repos.Account, deps.CBRService, deps.EmailService)
	analyticsService := NewAnalyticsService(deps.Repos.Transaction, deps.Repos.Credit, deps.Repos.Payment)

	return &Services{
		User:        userService,
		Account:     accountService,
		Card:        cardService,
		Transaction: transactionService,
		Credit:      creditService,
		Analytics:   analyticsService,
		CBR:         deps.CBRService,
		Email:       deps.EmailService,
		Encryption:  deps.EncryptionService,
	}
}
