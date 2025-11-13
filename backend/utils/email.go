package utils

import (
	"crypto/tls"
	"fmt"
	"friendship/config"
	"net/smtp"
)

func SendEmail(to, subject, body string, cfg *config.Config) error {
	from := cfg.Email.From
	password := cfg.Email.Password
	smtpHost := cfg.Email.SmtpHost
	smtpPort := cfg.Email.SmtpPort

	if smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP конфиг пустой: host='%s', port='%s'", smtpHost, smtpPort)
	}

	if from == "" || password == "" {
		return fmt.Errorf("SMTP данные")
	}

	message := []byte(fmt.Sprintf(
		"Subject: %s\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
			"%s", subject, from, to, body,
	))

	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	if smtpPort == "587" {
		return sendMailWithSTARTTLS(address, smtpHost, from, password, []string{to}, message)
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(address, auth, from, []string{to}, message)
}

func sendMailWithSTARTTLS(address, host, from, password string, to []string, msg []byte) error {
	client, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("не получилось подключиться: %w", err)
	}
	defer client.Close()

	if err = client.Hello(host); err != nil {
		return fmt.Errorf("проблема с EHLO: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName: host,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	auth := smtp.PlainAuth("", from, password, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
