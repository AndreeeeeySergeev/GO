package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// CreateHMAC создает HMAC для данных с использованием ключа
func CreateHMAC(data, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// VerifyHMAC проверяет HMAC для данных
func VerifyHMAC(data, signature, key string) error {
	expectedHMAC, err := CreateHMAC(data, key)
	if err != nil {
		return err
	}

	if expectedHMAC != signature {
		return ErrHMACVerificationFailed
	}

	return nil
}
