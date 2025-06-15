package repository

import (
	"database/sql"
	"errors"

	"bank-service/internal/models"
)

type AccountRepository interface {
	Create(account models.Account) (int64, error)
	GetByID(id int64) (models.Account, error)
	GetByNumber(number string) (models.Account, error)
	GetByUserID(userID int64) ([]models.Account, error)
	UpdateBalance(id int64, balance float64) error
	BeginTx() (*sql.Tx, error)
	UpdateBalanceTx(tx *sql.Tx, id int64, balance float64) error
}

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Create(account models.Account) (int64, error) {
	query := `
		INSERT INTO accounts (user_id, number, type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		account.UserID,
		account.Number,
		account.Type,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresAccountRepository) GetByID(id int64) (models.Account, error) {
	query := `
		SELECT id, user_id, number, type, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var account models.Account
	err := r.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Number,
		&account.Type,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Account{}, errors.New("account not found")
		}
		return models.Account{}, err
	}

	return account, nil
}

func (r *PostgresAccountRepository) GetByNumber(number string) (models.Account, error) {
	query := `
		SELECT id, user_id, number, type, balance, created_at, updated_at
		FROM accounts
		WHERE number = $1
	`

	var account models.Account
	err := r.db.QueryRow(query, number).Scan(
		&account.ID,
		&account.UserID,
		&account.Number,
		&account.Type,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Account{}, errors.New("account not found")
		}
		return models.Account{}, err
	}

	return account, nil
}

func (r *PostgresAccountRepository) GetByUserID(userID int64) ([]models.Account, error) {
	query := `
		SELECT id, user_id, number, type, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Number,
			&account.Type,
			&account.Balance,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *PostgresAccountRepository) UpdateBalance(id int64, balance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, balance, id)
	return err
}

func (r *PostgresAccountRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *PostgresAccountRepository) UpdateBalanceTx(tx *sql.Tx, id int64, balance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := tx.Exec(query, balance, id)
	return err
}
