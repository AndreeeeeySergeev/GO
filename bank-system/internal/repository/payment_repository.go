package repository

import (
	"database/sql"
	"time"

	"bank-service/internal/models"
)

type PaymentRepository interface {
	Create(payment models.PaymentSchedule) (int64, error)
	GetByCreditID(creditID int64) ([]models.PaymentSchedule, error)
	GetPendingPayments() ([]models.PaymentSchedule, error)
	UpdateStatus(id int64, status models.PaymentStatus, paidDate *time.Time) error
	CreateBatch(payments []models.PaymentSchedule) error
	CreateTx(tx *sql.Tx, payment models.PaymentSchedule) (int64, error)
	CreateBatchTx(tx *sql.Tx, payments []models.PaymentSchedule) error
}

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Create(payment models.PaymentSchedule) (int64, error) {
	query := `
		INSERT INTO payment_schedules (credit_id, payment_date, amount, principal, interest, remaining_debt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		payment.CreditID,
		payment.PaymentDate,
		payment.Amount,
		payment.Principal,
		payment.Interest,
		payment.RemainingDebt,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresPaymentRepository) GetByCreditID(creditID int64) ([]models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_date, amount, principal, interest, remaining_debt, status, paid_date, created_at, updated_at
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY payment_date
	`

	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.PaymentSchedule
	for rows.Next() {
		var payment models.PaymentSchedule
		var paidDate sql.NullTime

		if err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.PaymentDate,
			&payment.Amount,
			&payment.Principal,
			&payment.Interest,
			&payment.RemainingDebt,
			&payment.Status,
			&paidDate,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if paidDate.Valid {
			payment.PaidDate = &paidDate.Time
		}

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *PostgresPaymentRepository) GetPendingPayments() ([]models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_date, amount, principal, interest, remaining_debt, status, paid_date, created_at, updated_at
		FROM payment_schedules
		WHERE status = $1 AND payment_date <= NOW()
		ORDER BY payment_date
	`

	rows, err := r.db.Query(query, models.PaymentStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.PaymentSchedule
	for rows.Next() {
		var payment models.PaymentSchedule
		var paidDate sql.NullTime

		if err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.PaymentDate,
			&payment.Amount,
			&payment.Principal,
			&payment.Interest,
			&payment.RemainingDebt,
			&payment.Status,
			&paidDate,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if paidDate.Valid {
			payment.PaidDate = &paidDate.Time
		}

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *PostgresPaymentRepository) UpdateStatus(id int64, status models.PaymentStatus, paidDate *time.Time) error {
	var query string
	var args []interface{}

	if paidDate != nil {
		query = `
			UPDATE payment_schedules
			SET status = $1, paid_date = $2, updated_at = NOW()
			WHERE id = $3
		`
		args = []interface{}{status, paidDate, id}
	} else {
		query = `
			UPDATE payment_schedules
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		args = []interface{}{status, id}
	}

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *PostgresPaymentRepository) CreateBatch(payments []models.PaymentSchedule) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO payment_schedules (credit_id, payment_date, amount, principal, interest, remaining_debt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, payment := range payments {
		_, err := stmt.Exec(
			payment.CreditID,
			payment.PaymentDate,
			payment.Amount,
			payment.Principal,
			payment.Interest,
			payment.RemainingDebt,
			payment.Status,
			payment.CreatedAt,
			payment.UpdatedAt,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresPaymentRepository) CreateTx(tx *sql.Tx, payment models.PaymentSchedule) (int64, error) {
	query := `
		INSERT INTO payment_schedules (credit_id, payment_date, amount, principal, interest, remaining_debt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(
		query,
		payment.CreditID,
		payment.PaymentDate,
		payment.Amount,
		payment.Principal,
		payment.Interest,
		payment.RemainingDebt,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresPaymentRepository) CreateBatchTx(tx *sql.Tx, payments []models.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (credit_id, payment_date, amount, principal, interest, remaining_debt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, payment := range payments {
		_, err := stmt.Exec(
			payment.CreditID,
			payment.PaymentDate,
			payment.Amount,
			payment.Principal,
			payment.Interest,
			payment.RemainingDebt,
			payment.Status,
			payment.CreatedAt,
			payment.UpdatedAt,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
