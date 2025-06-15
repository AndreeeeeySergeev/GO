package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"bank-service/internal/models"
	"bank-service/internal/repository"
)

var (
	ErrUserExists         = errors.New("user with this email or username already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService interface {
	Register(input models.UserRegistration) (models.UserResponse, error)
	Login(input models.UserLogin) (string, error)
	GetByID(id int64) (models.UserResponse, error)
	ValidateToken(tokenString string) (int64, error)
}

type userService struct {
	repo       repository.UserRepository
	encryption EncryptionService
}

func NewUserService(repo repository.UserRepository, encryption EncryptionService) UserService {
	return &userService{
		repo:       repo,
		encryption: encryption,
	}
}

func (s *userService) Register(input models.UserRegistration) (models.UserResponse, error) {
	if err := input.Validate(); err != nil {
		return models.UserResponse{}, err
	}

	emailExists, err := s.repo.CheckEmailExists(input.Email)
	if err != nil {
		return models.UserResponse{}, err
	}

	if emailExists {
		return models.UserResponse{}, ErrUserExists
	}

	usernameExists, err := s.repo.CheckUsernameExists(input.Username)
	if err != nil {
		return models.UserResponse{}, err
	}

	if usernameExists {
		return models.UserResponse{}, ErrUserExists
	}

	passwordHash, err := s.encryption.HashPassword(input.Password)
	if err != nil {
		return models.UserResponse{}, err
	}

	now := time.Now()
	user := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
		FullName:     input.FullName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	id, err := s.repo.Create(user)
	if err != nil {
		return models.UserResponse{}, err
	}

	user.ID = id

	return models.ToUserResponse(user), nil
}

func (s *userService) Login(input models.UserLogin) (string, error) {
	user, err := s.repo.GetByEmail(input.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !s.encryption.CheckPasswordHash(input.Password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.encryption.GetJWTSecret()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) GetByID(id int64) (models.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return models.UserResponse{}, ErrUserNotFound
	}

	return models.ToUserResponse(user), nil
}

func (s *userService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.encryption.GetJWTSecret()), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid user_id in token")
		}
		return int64(userID), nil
	}

	return 0, errors.New("invalid token")
}
