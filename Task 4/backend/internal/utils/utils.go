package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func NormalizePhone(phone string) (string, error) {
	re := regexp.MustCompile(`\D`)
	cleaned := re.ReplaceAllString(phone, "")

	if len(cleaned) == 11 && strings.HasPrefix(cleaned, "8") {
		cleaned = "7" + cleaned[1:]
	}

	if len(cleaned) == 10 {
		cleaned = "7" + cleaned
	}

	if len(cleaned) != 11 || !strings.HasPrefix(cleaned, "7") {
		return "", errors.New("неверный формат телефона. Используйте формат: 7XXXXXXXXXX (11 цифр)")
	}

	return cleaned, nil
}

func GenerateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
