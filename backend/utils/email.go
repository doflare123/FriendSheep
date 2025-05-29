package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Формируем письмо
	message := []byte(fmt.Sprintf(
		"Subject: %s\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n"+
			"%s", subject, from, to, body,
	))

	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(address, auth, from, []string{to}, message)
}
