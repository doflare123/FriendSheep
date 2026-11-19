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
	"friendship/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type registerHandlerServiceStub struct {
	createSessionErr error
	createUserErr    error
	createUserInputs []registerservice.CreateUserInput
	createUserResult *dto.AuthResponse
}

func (s *registerHandlerServiceStub) CreateUser(_ context.Context, input registerservice.CreateUserInput) (*dto.AuthResponse, error) {
	s.createUserInputs = append(s.createUserInputs, input)
	if s.createUserErr != nil {
		return nil, s.createUserErr
	}
	if s.createUserResult != nil {
		return s.createUserResult, nil
	}
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

func TestRegisterHandlerCreateUserMapsOptionalProfileValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupRegisterHandlerValidation(t)

	for name, serviceErr := range map[string]error{
		"birthday": registerservice.ErrInvalidBirthday,
		"city ID":  registerservice.ErrInvalidCityID,
		"city":     registerservice.ErrCityNotFound,
	} {
		t.Run(name, func(t *testing.T) {
			h := handlers.NewRegisterHandler(&registerHandlerServiceStub{createUserErr: serviceErr})
			router := gin.New()
			router.POST("/api/v2/register/", h.CreateUser)

			body := `{"name":"Valid User","password":"Password123!","email":"user@example.com","session_id":"verified-session"}`
			req := httptest.NewRequest(http.MethodPost, "/api/v2/register/", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			assertCommonErrorShape(t, rec)
		})
	}
}

func TestRegisterHandlerCreateUserOptionalProfileJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupRegisterHandlerValidation(t)
	for _, tc := range []struct {
		name       string
		fields     string
		birthday   string
		cityID     uint
		wantStatus int
	}{
		{name: "omitted", wantStatus: http.StatusCreated},
		{name: "null", fields: `,"birthday":null,"cityId":null`, wantStatus: http.StatusCreated},
		{name: "both", fields: `,"birthday":"2000-02-29","cityId":7`, birthday: "2000-02-29", cityID: 7, wantStatus: http.StatusCreated},
		{name: "birthday only", fields: `,"birthday":"2001-02-03"`, birthday: "2001-02-03", wantStatus: http.StatusCreated},
		{name: "city only", fields: `,"cityId":7`, cityID: 7, wantStatus: http.StatusCreated},
		{name: "negative city", fields: `,"cityId":-1`, wantStatus: http.StatusBadRequest},
		{name: "string city", fields: `,"cityId":"7"`, wantStatus: http.StatusBadRequest},
		{name: "fractional city", fields: `,"cityId":1.5`, wantStatus: http.StatusBadRequest},
		{name: "numeric birthday", fields: `,"birthday":2000`, wantStatus: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &registerHandlerServiceStub{createUserResult: &dto.AuthResponse{AccessToken: "access-token"}}
			h := handlers.NewRegisterHandler(service)
			router := gin.New()
			router.POST("/api/v2/register/", h.CreateUser)
			body := `{"name":"Valid User","password":"Password123!","email":"user@example.com","session_id":"verified-session"` + tc.fields + `}`
			req := httptest.NewRequest(http.MethodPost, "/api/v2/register/", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.wantStatus == http.StatusBadRequest {
				if len(service.createUserInputs) != 0 {
					t.Fatal("service was called after JSON binding error")
				}
				assertCommonErrorShape(t, rec)
				return
			}
			if len(service.createUserInputs) != 1 {
				t.Fatalf("service calls = %d, want 1", len(service.createUserInputs))
			}
			input := service.createUserInputs[0]
			if tc.birthday == "" {
				if input.Birthday != nil {
					t.Fatalf("birthday = %v, want nil", input.Birthday)
				}
			} else if input.Birthday == nil || *input.Birthday != tc.birthday {
				t.Fatalf("birthday = %v, want %s", input.Birthday, tc.birthday)
			}
			if tc.cityID == 0 {
				if input.CityID != nil {
					t.Fatalf("city ID = %v, want nil", input.CityID)
				}
			} else if input.CityID == nil || *input.CityID != tc.cityID {
				t.Fatalf("city ID = %v, want %d", input.CityID, tc.cityID)
			}
		})
	}
}

func setupRegisterHandlerValidation(t *testing.T) {
	t.Helper()
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("unexpected Gin validator engine")
	}
	if err := v.RegisterValidation("username", utils.ValidateNameTag); err != nil {
		t.Fatalf("register username validator: %v", err)
	}
	if err := v.RegisterValidation("password", utils.PasswordValidation); err != nil {
		t.Fatalf("register password validator: %v", err)
	}
}
