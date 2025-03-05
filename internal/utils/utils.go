package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"unicode"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func GenerateID(s string, length int) (id string) {
	hash := sha256.Sum256([]byte(s))
	id = base64.RawURLEncoding.EncodeToString(hash[:length])[:length]
	return
}

func ValidateStruct(a any) error {
	return validator.New().Struct(a)
}

func ApplyLuhnAlgorithm(num int) string {
	var digits []int
	tempNum := num

	for tempNum > 0 {
		digits = append([]int{tempNum % 10}, digits...)
		tempNum /= 10
	}

	sum := 0
	nine := 9
	parity := (len(digits) + 1) % 2
	for i, d := range digits {
		if i%2 == parity {
			d *= 2
			if d > nine {
				d -= 9
			}
		}
		sum += d
	}

	checkDigit := (10 - (sum % 10)) % 10

	return strconv.Itoa(num) + strconv.Itoa(checkDigit)
}

func ValidateLuhnAlgorithm(number string) bool {
	nine := 9
	var sum int
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := number[i]

		if !unicode.IsDigit(rune(digit)) {
			continue
		}

		n := int(digit - '0')

		if alternate {
			n *= 2
			if n > nine {
				n -= 9
			}
		}

		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}
