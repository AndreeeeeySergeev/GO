package repository

import (
	"database/sql"
	"errors"

	"bank-service/internal/models"
)

type UserRepository interface {
	Create(user models.User) (int64, error)
	GetByID(id int64) (models.User, error)
	GetByEmail(email string) (models.User, error)
	GetByUsername(username string) (models.User, error)
	CheckEmailExists(email string) (bool, error)
	CheckUsernameExists(username string) (bool, error)
	Update(user models.User) error
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(user models.User) (int64, error) {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresUserRepository) GetByID(id int64) (models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(email string) (models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByUsername(username string) (models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) CheckEmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresUserRepository) CheckUsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresUserRepository) Update(user models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, full_name = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.UpdatedAt,
		user.ID,
	)

	return err
}
