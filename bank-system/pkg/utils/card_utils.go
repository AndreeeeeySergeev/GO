package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// Проверка номера карты по алгоритму Луна
func ValidateLuhn(cardNumber string) bool {
	var sum int
	var alternate bool

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n := int(cardNumber[i] - '0')
		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}

		sum += n
		alternate = !alternate
	}

	return sum%10 == 0
}

func GenerateCardNumber(prefix string) string {
	rand.Seed(time.Now().UnixNano())

	if prefix == "" {
		prefix = "4"
	}

	length := 16 - len(prefix)
	cardNumber := prefix

	for i := 0; i < length-1; i++ {
		cardNumber += fmt.Sprintf("%d", rand.Intn(10))
	}

	// Вычисление контрольной цифры по алгоритму Луна
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n := int(cardNumber[i] - '0')
		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}

		sum += n
		alternate = !alternate
	}

	checkDigit := (10 - (sum % 10)) % 10
	cardNumber += fmt.Sprintf("%d", checkDigit)

	return cardNumber
}
