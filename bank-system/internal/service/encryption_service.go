package service

import (
	"bank-service/internal/config"
	"bank-service/pkg/encryption"
)

type EncryptionService interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
	EncryptData(data string) (string, error)
	DecryptData(encryptedData string) (string, error)
	CreateHMAC(data string) (string, error)
	VerifyHMAC(data, signature string) error
	GetJWTSecret() string
}

type encryptionService struct {
	pgpKey    string
	hmacKey   string
	jwtSecret string
}

func NewEncryptionService(config *config.Config) EncryptionService {
	return &encryptionService{
		pgpKey:    config.Security.PGPKey,
		hmacKey:   config.Security.HMACKey,
		jwtSecret: config.Security.JWTSecret,
	}
}

func (s *encryptionService) HashPassword(password string) (string, error) {
	return encryption.HashPassword(password)
}

func (s *encryptionService) CheckPasswordHash(password, hash string) bool {
	return encryption.CheckPasswordHash(password, hash)
}

func (s *encryptionService) EncryptData(data string) (string, error) {
	return encryption.EncryptPGP(data, s.pgpKey)
}

func (s *encryptionService) DecryptData(encryptedData string) (string, error) {
	return encryption.DecryptPGP(encryptedData, s.pgpKey)
}

func (s *encryptionService) CreateHMAC(data string) (string, error) {
	return encryption.CreateHMAC(data, s.hmacKey)
}

func (s *encryptionService) VerifyHMAC(data, signature string) error {
	return encryption.VerifyHMAC(data, signature, s.hmacKey)
}

func (s *encryptionService) GetJWTSecret() string {
	return s.jwtSecret
}
