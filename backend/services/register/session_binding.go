package register

import (
	"errors"
	"fmt"
	"friendship/models"
	session "friendship/sessions"
	"net/mail"
	"strings"
)

func normalizeEmail(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", fmt.Errorf("Некорректный email: %w", err)
	}

	return addr.Address, nil
}

func validateVerifiedSessionBinding(
	sess *session.Session,
	expectedType models.SessionTypeReg,
	requestEmail string,
) (string, error) {
	if sess == nil {
		return "", ErrSessionNotFound
	}

	if sess.Type != expectedType {
		return "", ErrSessionTypeMismatch
	}

	if !sess.IsVerified {
		return "", ErrSessionNotVerified
	}

	sessionEmailRaw, ok := sess.Extra["email"]
	if !ok || strings.TrimSpace(sessionEmailRaw) == "" {
		return "", ErrSessionEmailMismatch
	}

	sessionEmail, err := normalizeEmail(sessionEmailRaw)
	if err != nil {
		return "", ErrSessionEmailMismatch
	}

	requestBoundEmail, err := normalizeEmail(requestEmail)
	if err != nil {
		return "", err
	}

	if !strings.EqualFold(sessionEmail, requestBoundEmail) {
		return "", ErrSessionEmailMismatch
	}

	return sessionEmail, nil
}

func mapRegisterSessionLookupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, session.ErrSessionNotFound):
		return ErrSessionNotFound
	default:
		return err
	}
}
