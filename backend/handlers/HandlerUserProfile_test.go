package handlers_test

import (
	"context"
	"errors"
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

type userProfileApplicationStub struct {
	viewerID  uint
	target    users.Target
	patch     users.PrivacyPatch
	tiles     users.TileSettings
	profiles  int
	privacies int
	patches   int
	err       error
}

func (s *userProfileApplicationStub) GetProfile(_ context.Context, viewerID uint, target users.Target) (users.Profile, error) {
	s.viewerID, s.target = viewerID, target
	s.profiles++
	return users.Profile{}, s.err
}

func (s *userProfileApplicationStub) GetPrivacy(_ context.Context, userID uint) (users.Privacy, error) {
	s.viewerID = userID
	s.privacies++
	return users.Privacy{}, s.err
}

func (s *userProfileApplicationStub) PatchPrivacy(_ context.Context, userID uint, patch users.PrivacyPatch) (users.Privacy, error) {
	s.viewerID, s.patch = userID, patch
	s.patches++
	return users.Privacy{}, s.err
}

func (s *userProfileApplicationStub) GetTileSettings(_ context.Context, userID uint) (users.TileSettings, error) {
	s.viewerID = userID
	return s.tiles, s.err
}

func (s *userProfileApplicationStub) ReplaceTileSettings(_ context.Context, userID uint, settings users.TileSettings) (users.TileSettings, error) {
	s.viewerID, s.tiles = userID, settings
	s.patches++
	return s.tiles, s.err
}

func TestUserProfileHandlerRejectsInvalidIdentifierBeforeService(t *testing.T) {
	application := &userProfileApplicationStub{}
	router := newUserProfileTestRouter(application, 7)

	for _, identifier := range []string{"0", "-1", "alice", "@"} {
		t.Run(identifier, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users/"+identifier, nil))
			if response.Code != http.StatusBadRequest || application.profiles != 0 {
				t.Fatalf("identifier %q response/calls = %d/%d, want 400/0; body=%s", identifier, response.Code, application.profiles, response.Body.String())
			}
		})
	}
}

func TestUserProfileHandlerUsesOnlyAuthenticatedViewerAndStrictPrivacyPatch(t *testing.T) {
	application := &userProfileApplicationStub{}
	router := newUserProfileTestRouter(application, 7)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users?userId=999", nil))
	if response.Code != http.StatusOK || application.profiles != 1 || application.viewerID != 7 || application.target != users.OwnTarget() {
		t.Fatalf("own profile response/call = %d/%#v, body=%s", response.Code, application, response.Body.String())
	}

	for _, body := range []string{"{}", `{"mode":null}`, `{"userId":999}`, `{"mode":"open","unknown":true}`, `{"visibleStatistics":null}`} {
		response = httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/privacy", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || application.patches != 0 {
			t.Fatalf("body %s response/calls = %d/%d, want 400/0; body=%s", body, response.Code, application.patches, response.Body.String())
		}
	}

	response = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/privacy", strings.NewReader(`{"mode":"custom","visibleStatistics":["completedEvents"]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.patches != 1 || application.viewerID != 7 || application.patch.Mode == nil || application.patch.VisibleStatistics == nil {
		t.Fatalf("valid patch response/call = %d/%#v, body=%s", response.Code, application, response.Body.String())
	}
}

func TestUserProfileHandlerReplacesTileSettingsStrictly(t *testing.T) {
	application := &userProfileApplicationStub{}
	router := newUserProfileTestRouter(application, 7)
	for _, body := range []string{"{}", `{"tiles":null}`, `{"tiles":[],"userId":9}`} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/statistics/tiles", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || application.patches != 0 {
			t.Fatalf("body %s response/calls = %d/%d, want 400/0; body=%s", body, response.Code, application.patches, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/statistics/tiles", strings.NewReader(`{"tiles":["games","films"]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.patches != 1 || application.viewerID != 7 || len(application.tiles.Tiles) != 2 || application.tiles.Tiles[0] != users.TileGames || application.tiles.Tiles[1] != users.TileFilms {
		t.Fatalf("valid replacement response/application = %d/%#v, body=%s", response.Code, application, response.Body.String())
	}

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users/me/statistics/tiles?userId=9", nil))
	if response.Code != http.StatusOK || application.viewerID != 7 {
		t.Fatalf("tile settings response/application = %d/%#v", response.Code, application)
	}
}

func TestUserProfileRoutesRequireAuthenticationAndKeepPrivacyStaticRoute(t *testing.T) {
	application := &userProfileApplicationStub{}
	router := gin.New()
	routes.RegisterUserProfileRoutes(router, handlers.NewUserProfileHandler(application), func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	for _, path := range []string{"/api/v2/users", "/api/v2/users/17", "/api/v2/users/@123", "/api/v2/users/me/privacy"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s status = %d, want 401", path, response.Code)
		}
	}
	if application.profiles != 0 || application.privacies != 0 || application.patches != 0 {
		t.Fatalf("application was called despite rejected auth: %#v", application)
	}
}

func TestUserProfileHandlerMapsClosedProfile(t *testing.T) {
	application := &userProfileApplicationStub{err: users.ErrProfileClosed}
	router := newUserProfileTestRouter(application, 7)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users/8", nil))
	if response.Code != http.StatusForbidden || !errors.Is(application.err, users.ErrProfileClosed) || application.profiles != 1 {
		t.Fatalf("response/application = %d/%#v, body=%s", response.Code, application, response.Body.String())
	}
}

type profileAuthSessionsStub struct{ active bool }

func (s profileAuthSessionsStub) HasActiveSession(_ context.Context, _ string) (bool, error) {
	return s.active, nil
}

func TestUserProfileRoutesUseRealJWTAlongsideUsersSiblingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := utils.NewJWTUtils("test-secret")
	activeAuth := middlewares.NewAuthMiddleware(jwtService, profileAuthSessionsStub{active: true})
	revokedAuth := middlewares.NewAuthMiddleware(jwtService, profileAuthSessionsStub{active: false})
	application := &userProfileApplicationStub{}

	newRouter := func(auth *middlewares.AuthMiddleware) *gin.Engine {
		router := gin.New()
		routes.RegisterUserProfileRoutes(router, handlers.NewUserProfileHandler(application), auth.RequireAuth())
		routes.RegisterNotificationRoutes(router, handlers.NewNotificationsHandler(&notificationInboxStub{}), auth.RequireAuth())
		routes.RegisterStatisticsRoutes(router, handlers.NewStatisticsHandler(&statisticsApplicationStub{}, nil), auth.RequireAuth())
		return router
	}
	paths := []struct{ method, path string }{
		{http.MethodGet, "/api/v2/users"},
		{http.MethodGet, "/api/v2/users/7"},
		{http.MethodGet, "/api/v2/users/@alice"},
		{http.MethodGet, "/api/v2/users/me/privacy"},
		{http.MethodPatch, "/api/v2/users/me/privacy"},
	}
	for _, authCase := range []struct {
		name, token string
		router      *gin.Engine
	}{
		{name: "missing", router: newRouter(activeAuth)},
		{name: "invalid", token: "invalid", router: newRouter(activeAuth)},
		{name: "revoked", token: mustProfileAccessToken(t, jwtService, 7), router: newRouter(revokedAuth)},
	} {
		for _, route := range paths {
			t.Run(authCase.name+route.method+route.path, func(t *testing.T) {
				request := httptest.NewRequest(route.method, route.path, strings.NewReader(`{"mode":"open"}`))
				if authCase.token != "" {
					request.Header.Set("Authorization", "Bearer "+authCase.token)
				}
				response := httptest.NewRecorder()
				authCase.router.ServeHTTP(response, request)
				if response.Code != http.StatusUnauthorized {
					t.Fatalf("status = %d, want 401; body=%s", response.Code, response.Body.String())
				}
			})
		}
	}

	validRouter := newRouter(activeAuth)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v2/users", nil)
	request.Header.Set("Authorization", "Bearer "+mustProfileAccessToken(t, jwtService, 7))
	validRouter.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.profiles != 1 || application.viewerID != 7 {
		t.Fatalf("valid request status/application = %d/%#v; body=%s", response.Code, application, response.Body.String())
	}
}

func mustProfileAccessToken(t *testing.T, jwtService *utils.JWTUtils, userID uint) string {
	t.Helper()
	token, _, _, err := jwtService.GenerateAccessToken(userID, "profile-session")
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func newUserProfileTestRouter(application handlers.UserProfileApplication, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.RegisterUserProfileRoutes(router, handlers.NewUserProfileHandler(application), func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	return router
}

var _ handlers.UserProfileApplication = (*userProfileApplicationStub)(nil)
