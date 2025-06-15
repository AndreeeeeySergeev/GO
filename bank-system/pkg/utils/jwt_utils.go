package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Генерация JWT токена
func GenerateJWT(userID int64, secret string, expirationTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expirationTime).Unix(),
	})

	// Подписание токена секретным ключом
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Проверка JWT токена и возвращение ID пользователя
func ValidateJWT(tokenString string, secret string) (int64, error) {
	// Парсинг токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	// Проверка валидности токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Извлечение ID пользователя
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid user_id in token")
		}

		return int64(userID), nil
	}

	return 0, errors.New("invalid token")
}
