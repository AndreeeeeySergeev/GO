package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"bank-service/internal/config"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

type Repositories struct {
	User        UserRepository
	Account     AccountRepository
	Card        CardRepository
	Transaction TransactionRepository
	Credit      CreditRepository
	Payment     PaymentRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:        NewUserRepository(db),
		Account:     NewAccountRepository(db),
		Card:        NewCardRepository(db),
		Transaction: NewTransactionRepository(db),
		Credit:      NewCreditRepository(db),
		Payment:     NewPaymentRepository(db),
	}
}
