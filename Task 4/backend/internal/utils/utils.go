package utils

import (
	"errors"
	"regexp"
	"strings"
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
