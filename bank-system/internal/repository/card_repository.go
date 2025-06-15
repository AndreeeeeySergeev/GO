package repository

import (
	"database/sql"
	"errors"

	"bank-service/internal/models"
)

type CardRepository interface {
	Create(card models.Card) (int64, error)
	GetByID(id int64) (models.Card, error)
	GetByAccountID(accountID int64) ([]models.Card, error)
	GetByUserID(userID int64) ([]models.Card, error)
	UpdateStatus(id int64, isActive bool) error
}

type PostgresCardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &PostgresCardRepository{db: db}
}

func (r *PostgresCardRepository) Create(card models.Card) (int64, error) {
	query := `
		INSERT INTO cards (account_id, user_id, number_encrypted, number_hmac, expiry_date_encrypted, 
		                  expiry_date_hmac, cvv_hash, type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		card.AccountID,
		card.UserID,
		card.Number,
		card.NumberHMAC,
		card.ExpiryDate,
		card.ExpiryHMAC,
		card.CVV,
		card.Type,
		card.IsActive,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresCardRepository) GetByID(id int64) (models.Card, error) {
	query := `
		SELECT id, account_id, user_id, number_encrypted, number_hmac, expiry_date_encrypted, 
		       expiry_date_hmac, cvv_hash, type, is_active, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	var card models.Card
	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.AccountID,
		&card.UserID,
		&card.Number,
		&card.NumberHMAC,
		&card.ExpiryDate,
		&card.ExpiryHMAC,
		&card.CVV,
		&card.Type,
		&card.IsActive,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Card{}, errors.New("card not found")
		}
		return models.Card{}, err
	}

	return card, nil
}

func (r *PostgresCardRepository) GetByAccountID(accountID int64) ([]models.Card, error) {
	query := `
		SELECT id, account_id, user_id, number_encrypted, number_hmac, expiry_date_encrypted, 
		       expiry_date_hmac, cvv_hash, type, is_active, created_at, updated_at
		FROM cards
		WHERE account_id = $1
	`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.UserID,
			&card.Number,
			&card.NumberHMAC,
			&card.ExpiryDate,
			&card.ExpiryHMAC,
			&card.CVV,
			&card.Type,
			&card.IsActive,
			&card.CreatedAt,
			&card.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (r *PostgresCardRepository) GetByUserID(userID int64) ([]models.Card, error) {
	query := `
		SELECT id, account_id, user_id, number_encrypted, number_hmac, expiry_date_encrypted, 
		       expiry_date_hmac, cvv_hash, type, is_active, created_at, updated_at
		FROM cards
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.UserID,
			&card.Number,
			&card.NumberHMAC,
			&card.ExpiryDate,
			&card.ExpiryHMAC,
			&card.CVV,
			&card.Type,
			&card.IsActive,
			&card.CreatedAt,
			&card.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (r *PostgresCardRepository) UpdateStatus(id int64, isActive bool) error {
	query := `
		UPDATE cards
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, isActive, id)
	return err
}
