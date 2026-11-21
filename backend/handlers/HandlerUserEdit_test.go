package handlers_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services/users"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type userEditApplicationStub struct {
	id      uint
	patch   users.ProfilePatch
	patches int
	uploads int
	deletes int
	err     error
}

func (s *userEditApplicationStub) GetOwn(_ context.Context, userID uint) (users.EditableProfile, error) {
	s.id = userID
	return users.EditableProfile{ID: userID}, s.err
}
func (s *userEditApplicationStub) Patch(_ context.Context, userID uint, patch users.ProfilePatch) (users.EditableProfile, error) {
	s.id, s.patch, s.patches = userID, patch, s.patches+1
	return users.EditableProfile{ID: userID}, s.err
}
func (s *userEditApplicationStub) UploadImage(_ context.Context, userID uint, _ io.Reader) (users.EditableProfile, error) {
	s.id, s.uploads = userID, s.uploads+1
	return users.EditableProfile{ID: userID}, s.err
}
func (s *userEditApplicationStub) DeleteImage(_ context.Context, userID uint) (users.EditableProfile, error) {
	s.id, s.deletes = userID, s.deletes+1
	return users.EditableProfile{ID: userID}, s.err
}

func TestUserEditHandlerRejectsMalformedPatchBeforeApplication(t *testing.T) {
	app := &userEditApplicationStub{}
	router := newUserEditTestRouter(app)
	for _, body := range []string{"{}", `{"name":null}`, `{"name":"a","name":"b"}`, `{"unknown":true}`, `{"unknown":null}`, `{"birthday":"20-01-01"}`, `{"cityId":"7"}`} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || app.patches != 0 {
			t.Fatalf("body %s: status/calls = %d/%d; body=%s", body, response.Code, app.patches, response.Body.String())
		}
	}
}

func TestUserEditHandlerUsesAuthenticatedUserAndMapsErrors(t *testing.T) {
	app := &userEditApplicationStub{}
	router := newUserEditTestRouter(app)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile?userId=99", bytes.NewBufferString(`{"us":"@Alice_2","birthday":null,"cityId":null}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || app.id != 7 || app.patches != 1 || app.patch.Us == nil || *app.patch.Us != "@Alice_2" || !app.patch.Birthday.Set || app.patch.Birthday.Value != nil || !app.patch.CityID.Set || app.patch.CityID.Value != nil {
		t.Fatalf("status/application = %d/%#v; body=%s", response.Code, app, response.Body.String())
	}
	app.err = users.ErrUsTaken
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile", bytes.NewBufferString(`{"name":"Alice"}`))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("us conflict status = %d; body=%s", response.Code, response.Body.String())
	}
	app.err = users.ErrCityNotFound
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile", bytes.NewBufferString(`{"cityId":777}`))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("city error status = %d; body=%s", response.Code, response.Body.String())
	}
	app.err = users.ErrBirthdayChangeDenied
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile", bytes.NewBufferString(`{"birthday":"2000-01-02"}`))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"error":"birthday_change_denied"`) {
		t.Fatalf("birthday denial status/body = %d/%s", response.Code, response.Body.String())
	}
	app.err = nil
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/profile", bytes.NewBufferString(`{"status":"`+strings.Repeat("x", 9000)+`"}`))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("large patch status = %d; body=%s", response.Code, response.Body.String())
	}
}

func TestUserEditRoutesRequireRealJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := utils.NewJWTUtils("edit-secret")
	auth := middlewares.NewAuthMiddleware(jwtService, profileAuthSessionsStub{active: true})
	app := &userEditApplicationStub{}
	profileApp := &userProfileApplicationStub{}
	router := gin.New()
	routes.RegisterUserProfileRoutes(router, handlers.NewUserProfileHandler(profileApp), auth.RequireAuth())
	routes.RegisterUserEditRoutes(router, handlers.NewUserEditHandler(app), auth.RequireAuth())
	for _, method := range []string{http.MethodGet, http.MethodPatch, http.MethodPut, http.MethodDelete} {
		path := "/api/v2/users/me/profile"
		if method == http.MethodPut || method == http.MethodDelete {
			path += "/image"
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(method, path, nil))
		if response.Code != http.StatusUnauthorized || app.id != 0 {
			t.Fatalf("missing token %s status/application = %d/%#v", method, response.Code, app)
		}
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v2/users/me/profile", nil)
	request.Header.Set("Authorization", "Bearer "+mustProfileAccessToken(t, jwtService, 7))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || app.id != 7 {
		t.Fatalf("valid route status/application = %d/%#v", response.Code, app)
	}
}

func TestUserEditHandlerImageRequiresExactlyOneFileAndMapsStorageFailure(t *testing.T) {
	app := &userEditApplicationStub{}
	router := newUserEditTestRouter(app)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/api/v2/users/me/profile/image", nil))
	if response.Code != http.StatusBadRequest || app.uploads != 0 {
		t.Fatalf("empty multipart status/calls = %d/%d", response.Code, app.uploads)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write([]byte("not validated by handler")); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	newRequest := func() *http.Request {
		request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/profile/image", bytes.NewReader(body.Bytes()))
		request.Header.Set("Content-Type", writer.FormDataContentType())
		return request
	}
	response = httptest.NewRecorder()
	router.ServeHTTP(response, newRequest())
	if response.Code != http.StatusOK || app.uploads != 1 || app.id != 7 {
		t.Fatalf("valid form status/application = %d/%#v; body=%s", response.Code, app, response.Body.String())
	}
	app.err = users.ErrImageStorageUnavailable
	response = httptest.NewRecorder()
	router.ServeHTTP(response, newRequest())
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("storage error status = %d; body=%s", response.Code, response.Body.String())
	}
}

func newUserEditTestRouter(app handlers.UserEditApplication) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v2/users")
	group.Use(func(c *gin.Context) { c.Set("userID", uint(7)); c.Next() })
	h := handlers.NewUserEditHandler(app)
	group.GET("/me/profile", h.GetOwn)
	group.PATCH("/me/profile", h.Patch)
	group.PUT("/me/profile/image", h.UploadImage)
	group.DELETE("/me/profile/image", h.DeleteImage)
	return router
}

var _ handlers.UserEditApplication = (*userEditApplicationStub)(nil)
