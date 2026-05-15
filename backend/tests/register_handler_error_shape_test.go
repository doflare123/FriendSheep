package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/models"
	"friendship/models/dto"
	registerservice "friendship/services/register"

	"github.com/gin-gonic/gin"
)

type registerHandlerServiceStub struct {
	createSessionErr error
}

func (s *registerHandlerServiceStub) CreateUser(context.Context, registerservice.CreateUserInput) (*dto.AuthResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *registerHandlerServiceStub) CreateSessionRegister(context.Context, string, string) (*models.SessionRegResponse, error) {
	if s.createSessionErr != nil {
		return nil, s.createSessionErr
	}
	return &models.SessionRegResponse{SessionID: "ok"}, nil
}

func (s *registerHandlerServiceStub) VerifySession(context.Context, registerservice.VerifySessionInput) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *registerHandlerServiceStub) ChangePassword(context.Context, registerservice.ChangePasswordInput) error {
	return errors.New("not implemented")
}

func TestRegisterHandlerCreateSessionRegisterInternalErrorUsesCommonShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handlers.NewRegisterHandler(&registerHandlerServiceStub{
		createSessionErr: errors.New("smtp down"),
	})
	router := gin.New()
	router.POST("/api/v2/register/session/register", h.CreateSessionRegister)

	body := `{"email":"user@example.com","type_ses":"register"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/register/session/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	assertCommonErrorShape(t, rec)
}

func TestRegisterHandlerVerifySessionBindErrorUsesCommonShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handlers.NewRegisterHandler(&registerHandlerServiceStub{})
	router := gin.New()
	router.PATCH("/api/v2/register/session/verify", h.VerifySession)

	req := httptest.NewRequest(http.MethodPatch, "/api/v2/register/session/verify", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertCommonErrorShape(t, rec)
}
