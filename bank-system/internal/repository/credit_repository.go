package repository

import (
	"database/sql"
	"errors"

	"bank-service/internal/models"
)

type CreditRepository interface {
	Create(credit models.Credit) (int64, error)
	GetByID(id int64) (models.Credit, error)
	GetByUserID(userID int64) ([]models.Credit, error)
	GetActiveCredits() ([]models.Credit, error)
	UpdateStatus(id int64, status models.CreditStatus) error
	BeginTx() (*sql.Tx, error)
	CreateTx(tx *sql.Tx, credit models.Credit) (int64, error)
}

type PostgresCreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) CreditRepository {
	return &PostgresCreditRepository{db: db}
}

func (r *PostgresCreditRepository) Create(credit models.Credit) (int64, error) {
	query := `
		INSERT INTO credits (user_id, account_id, amount, term, interest_rate, monthly_payment, total_payment, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.Term,
		credit.InterestRate,
		credit.MonthlyPayment,
		credit.TotalPayment,
		credit.Status,
		credit.StartDate,
		credit.EndDate,
		credit.CreatedAt,
		credit.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresCreditRepository) GetByID(id int64) (models.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, term, interest_rate, monthly_payment, total_payment, status, start_date, end_date, created_at, updated_at
		FROM credits
		WHERE id = $1
	`

	var credit models.Credit
	err := r.db.QueryRow(query, id).Scan(
		&credit.ID,
		&credit.UserID,
		&credit.AccountID,
		&credit.Amount,
		&credit.Term,
		&credit.InterestRate,
		&credit.MonthlyPayment,
		&credit.TotalPayment,
		&credit.Status,
		&credit.StartDate,
		&credit.EndDate,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Credit{}, errors.New("credit not found")
		}
		return models.Credit{}, err
	}

	return credit, nil
}

func (r *PostgresCreditRepository) GetByUserID(userID int64) ([]models.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, term, interest_rate, monthly_payment, total_payment, status, start_date, end_date, created_at, updated_at
		FROM credits
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []models.Credit
	for rows.Next() {
		var credit models.Credit
		if err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.Term,
			&credit.InterestRate,
			&credit.MonthlyPayment,
			&credit.TotalPayment,
			&credit.Status,
			&credit.StartDate,
			&credit.EndDate,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		); err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credits, nil
}

func (r *PostgresCreditRepository) GetActiveCredits() ([]models.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, term, interest_rate, monthly_payment, total_payment, status, start_date, end_date, created_at, updated_at
		FROM credits
		WHERE status IN ($1, $2)
	`

	rows, err := r.db.Query(query, models.CreditStatusActive, models.CreditStatusOverdue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []models.Credit
	for rows.Next() {
		var credit models.Credit
		if err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.Term,
			&credit.InterestRate,
			&credit.MonthlyPayment,
			&credit.TotalPayment,
			&credit.Status,
			&credit.StartDate,
			&credit.EndDate,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		); err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credits, nil
}

func (r *PostgresCreditRepository) UpdateStatus(id int64, status models.CreditStatus) error {
	query := `
		UPDATE credits
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *PostgresCreditRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *PostgresCreditRepository) CreateTx(tx *sql.Tx, credit models.Credit) (int64, error) {
	query := `
		INSERT INTO credits (user_id, account_id, amount, term, interest_rate, monthly_payment, total_payment, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(
		query,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.Term,
		credit.InterestRate,
		credit.MonthlyPayment,
		credit.TotalPayment,
		credit.Status,
		credit.StartDate,
		credit.EndDate,
		credit.CreatedAt,
		credit.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
