package repository

import (
	"database/sql"
	"time"

	"bank-service/internal/models"
)

type TransactionRepository interface {
	Create(transaction models.Transaction) (int64, error)
	GetByID(id int64) (models.Transaction, error)
	GetByUserID(userID int64, limit, offset int) ([]models.Transaction, error)
	GetByAccountID(accountID int64, limit, offset int) ([]models.Transaction, error)
	GetUserTransactionsByPeriod(userID int64, startDate, endDate time.Time) ([]models.Transaction, error)
	CreateTx(tx *sql.Tx, transaction models.Transaction) (int64, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(transaction models.Transaction) (int64, error) {
	query := `
		INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		transaction.UserID,
		transaction.FromAccountID,
		transaction.ToAccountID,
		transaction.Type,
		transaction.Amount,
		transaction.Description,
		transaction.Status,
		transaction.TransactionDate,
		transaction.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresTransactionRepository) GetByID(id int64) (models.Transaction, error) {
	query := `
		SELECT id, user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at
		FROM transactions
		WHERE id = $1
	`

	var transaction models.Transaction
	var fromAccountID, toAccountID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&fromAccountID,
		&toAccountID,
		&transaction.Type,
		&transaction.Amount,
		&transaction.Description,
		&transaction.Status,
		&transaction.TransactionDate,
		&transaction.CreatedAt,
	)

	if err != nil {
		return models.Transaction{}, err
	}

	if fromAccountID.Valid {
		transaction.FromAccountID = &fromAccountID.Int64
	}

	if toAccountID.Valid {
		transaction.ToAccountID = &toAccountID.Int64
	}

	return transaction, nil
}

func (r *PostgresTransactionRepository) GetByUserID(userID int64, limit, offset int) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY transaction_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var fromAccountID, toAccountID sql.NullInt64

		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&fromAccountID,
			&toAccountID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Status,
			&transaction.TransactionDate,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fromAccountID.Valid {
			transaction.FromAccountID = &fromAccountID.Int64
		}

		if toAccountID.Valid {
			transaction.ToAccountID = &toAccountID.Int64
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *PostgresTransactionRepository) GetByAccountID(accountID int64, limit, offset int) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at
		FROM transactions
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY transaction_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var fromAccountID, toAccountID sql.NullInt64

		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&fromAccountID,
			&toAccountID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Status,
			&transaction.TransactionDate,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fromAccountID.Valid {
			transaction.FromAccountID = &fromAccountID.Int64
		}

		if toAccountID.Valid {
			transaction.ToAccountID = &toAccountID.Int64
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *PostgresTransactionRepository) GetUserTransactionsByPeriod(userID int64, startDate, endDate time.Time) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at
		FROM transactions
		WHERE user_id = $1 AND transaction_date BETWEEN $2 AND $3
		ORDER BY transaction_date
	`

	rows, err := r.db.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var fromAccountID, toAccountID sql.NullInt64

		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&fromAccountID,
			&toAccountID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Status,
			&transaction.TransactionDate,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}

		if fromAccountID.Valid {
			transaction.FromAccountID = &fromAccountID.Int64
		}

		if toAccountID.Valid {
			transaction.ToAccountID = &toAccountID.Int64
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *PostgresTransactionRepository) CreateTx(tx *sql.Tx, transaction models.Transaction) (int64, error) {
	query := `
		INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description, status, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(
		query,
		transaction.UserID,
		transaction.FromAccountID,
		transaction.ToAccountID,
		transaction.Type,
		transaction.Amount,
		transaction.Description,
		transaction.Status,
		transaction.TransactionDate,
		transaction.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
