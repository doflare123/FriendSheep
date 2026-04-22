package utils

import (
	"crypto/tls"
	"fmt"
	"friendship/config"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const smtpTimeout = 15 * time.Second

func SendEmail(to, subject, body string, cfg *config.Config) error {
	from := strings.TrimSpace(cfg.Email.From)
	to = strings.TrimSpace(to)
	username := strings.TrimSpace(cfg.Email.Username)
	password := cfg.Email.Password
	smtpHost := cfg.Email.SmtpHost
	smtpPort := cfg.Email.SmtpPort

	if smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP config is empty: host=%q, port=%q", smtpHost, smtpPort)
	}

	if from == "" || password == "" {
		return fmt.Errorf("SMTP credentials are empty")
	}

	fromAddr, err := parseEmailAddress(from, "SMTP sender")
	if err != nil {
		return err
	}
	toAddr, err := parseEmailAddress(to, "recipient")
	if err != nil {
		return err
	}
	if fromAddr.Name == "" {
		fromAddr.Name = "FriendSheep"
	}
	if username == "" {
		username = fromAddr.Address
	}

	message := []byte(fmt.Sprintf(
		"Date: %s\r\n"+
			"Subject: %s\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"Content-Transfer-Encoding: 8bit\r\n\r\n"+
			"%s",
		time.Now().Format(time.RFC1123Z),
		mime.QEncoding.Encode("UTF-8", subject),
		fromAddr.String(),
		toAddr.String(),
		body,
	))

	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	switch smtpPort {
	case "465":
		return sendMailWithTLS(address, smtpHost, username, fromAddr.Address, password, []string{toAddr.Address}, message)
	case "587":
		return sendMailWithSTARTTLS(address, smtpHost, username, fromAddr.Address, password, []string{toAddr.Address}, message)
	default:
		auth := smtp.PlainAuth("", username, password, smtpHost)
		return smtp.SendMail(address, auth, fromAddr.Address, []string{toAddr.Address}, message)
	}
}

func parseEmailAddress(value, field string) (*mail.Address, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid email address: %w", field, err)
	}
	if !strings.Contains(addr.Address, "@") {
		return nil, fmt.Errorf("%s must include @domain: %q", field, value)
	}
	return addr, nil
}

func sendMailWithSTARTTLS(address, host, username, from, password string, to []string, msg []byte) error {
	dialer := &net.Dialer{Timeout: smtpTimeout}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("SMTP connect failed (%s): %w", address, err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("SMTP greeting failed (%s): %w", address, err)
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP EHLO failed: %w", err)
	}

	if err = client.StartTLS(&tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: host,
	}); err != nil {
		return fmt.Errorf("SMTP STARTTLS failed: %w", err)
	}

	return sendSMTPMessage(client, host, username, from, password, to, msg)
}

func sendMailWithTLS(address, host, username, from, password string, to []string, msg []byte) error {
	dialer := &net.Dialer{Timeout: smtpTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: host,
	})
	if err != nil {
		return fmt.Errorf("SMTP TLS connect failed (%s): %w", address, err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("SMTP TLS greeting failed (%s): %w", address, err)
	}
	defer client.Close()

	return sendSMTPMessage(client, host, username, from, password, to, msg)
}

func sendSMTPMessage(client *smtp.Client, host, username, from, password string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP sender rejected: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP recipient rejected (%s): %w", addr, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	if _, err = writer.Write(msg); err != nil {
		return fmt.Errorf("SMTP message write failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("SMTP message close failed: %w", err)
	}

	return client.Quit()
}
