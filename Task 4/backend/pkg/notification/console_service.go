package notification

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

type ConsoleService struct{}

func NewConsoleService() *ConsoleService {
	return &ConsoleService{}
}

func (c *ConsoleService) SendSMS(phone, code, purpose string) error {
	if !isValidPhone(phone) {
		return fmt.Errorf("invalid phone format")
	}

	log.Printf("\nSMS УВЕДОМЛЕНИЕ")
	log.Printf("Телефон: %s", phone)
	log.Printf("Код: %s", code)
	log.Printf("Цель: %s", purpose)
	log.Printf("Время: %s", time.Now().Format("15:04:05"))
	log.Println("─────────────────────")

	fmt.Printf("\n [SMS] Код для %s: %s (%s)\n", phone, code, purpose)
	return nil
}

func (c *ConsoleService) SendEmail(email, code, purpose string) error {
	if !isValidEmail(email) {
		return fmt.Errorf("invalid email format")
	}

	log.Printf("\nEMAIL УВЕДОМЛЕНИЕ")
	log.Printf("Email: %s", email)
	log.Printf("Код: %s", code)
	log.Printf("Цель: %s", purpose)
	log.Printf("Время: %s", time.Now().Format("15:04:05"))
	log.Println("─────────────────────")

	fmt.Printf("\n[EMAIL] Код для %s: %s (%s)\n", email, code, purpose)
	return nil
}

func (c *ConsoleService) VerifySMS(phone, code, purpose string) error {
	fmt.Printf("\n[SMS] Проверка кода для %s: %s (%s)\n", phone, code, purpose)
	return nil
}

func (c *ConsoleService) VerifyEmail(email, code, purpose string) error {
	fmt.Printf("\n[EMAIL] Проверка кода для %s: %s (%s)\n", email, code, purpose)
	return nil
}

func isValidPhone(phone string) bool {
	regex := regexp.MustCompile(`^\+7\d{10}$`)
	return regex.MatchString(phone)
}

func isValidEmail(email string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(email)
}
