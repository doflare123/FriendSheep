package tests

import (
	"testing"

	"friendship/config"
	"friendship/utils"
)

func TestSendEmailRejectsSenderWithoutAt(t *testing.T) {
	err := utils.SendEmail("user@example.com", "Subject", "Body", emailConfig("noreply"))
	if err == nil {
		t.Fatal("SendEmail returned nil error")
	}
}

func TestSendEmailRejectsRecipientWithoutAt(t *testing.T) {
	err := utils.SendEmail("drum2858gmail.com", "Subject", "Body", emailConfig("noreply@example.com"))
	if err == nil {
		t.Fatal("SendEmail returned nil error")
	}
}

func emailConfig(from string) *config.Config {
	return &config.Config{
		Email: config.EmailConfig{
			From:     from,
			Username: "smtp-login",
			Password: "password",
			SmtpHost: "smtp.example.com",
			SmtpPort: "587",
		},
	}
}
