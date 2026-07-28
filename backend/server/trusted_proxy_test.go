package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConfigureTrustedProxiesControlsGinClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		trustedProxies []string
		remoteAddr     string
		forwardedFor   string
		wantClientIP   string
	}{
		{
			name:           "default trusts no proxy",
			trustedProxies: nil,
			remoteAddr:     "10.1.2.3:1234",
			forwardedFor:   "203.0.113.90",
			wantClientIP:   "10.1.2.3",
		},
		{
			name:           "untrusted peer cannot forward",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "192.0.2.72:1234",
			forwardedFor:   "203.0.113.90",
			wantClientIP:   "192.0.2.72",
		},
		{
			name:           "configured proxy can forward",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "10.1.2.3:1234",
			forwardedFor:   "203.0.113.90",
			wantClientIP:   "203.0.113.90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			if err := configureTrustedProxies(router, tt.trustedProxies); err != nil {
				t.Fatalf("configureTrustedProxies(%#v): %v", tt.trustedProxies, err)
			}
			router.GET("/client-ip", func(c *gin.Context) {
				c.Header("X-Test-Client-IP", c.ClientIP())
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/client-ip", nil)
			request.RemoteAddr = tt.remoteAddr
			request.Header.Set("X-Forwarded-For", tt.forwardedFor)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
			}
			if got := recorder.Header().Get("X-Test-Client-IP"); got != tt.wantClientIP {
				t.Fatalf("Gin ClientIP = %q, want %q", got, tt.wantClientIP)
			}
		})
	}
}

func TestConfigureTrustedProxiesRejectsInvalidNetwork(t *testing.T) {
	router := gin.New()

	err := configureTrustedProxies(router, []string{"not-a-proxy"})
	if err == nil {
		t.Fatal("configureTrustedProxies() error = nil for invalid proxy")
	}
}
