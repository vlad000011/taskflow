package middleware

import (
	"net/http/httptest"
	"testing"

	"taskflow/backend/internal/config"

	"github.com/gin-gonic/gin"
)

func TestRolesMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, role string
		allowed    []string
		want       int
	}{
		{"admin-allowed", "admin", []string{"admin"}, 204},
		{"participant-denied", "participant", []string{"admin"}, 403},
		{"teacher-denied", "teacher", []string{"lead", "admin"}, 403},
		{"lead-allowed", "lead", []string{"lead", "admin"}, 204},
		{"missing-role-denied", "", []string{"admin"}, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.role != "" {
					c.Set("role", tc.role)
				}
				c.Next()
			})
			r.GET("/protected", Roles(tc.allowed...), func(c *gin.Context) { c.Status(204) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/protected", nil))
			if w.Code != tc.want {
				t.Fatalf("status=%d, want %d; body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestAuthMiddlewareRejectsMissingOrInvalidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// A nil DB is safe here: malformed/missing credentials must be rejected
	// before the middleware attempts to load a user from persistence.
	r := gin.New()
	r.Use(Auth(config.Config{JWTSecret: "test-secret-with-at-least-32-characters"}, nil))
	r.GET("/protected", func(c *gin.Context) { c.Status(204) })
	for _, tc := range []struct{ name, authorization string }{
		{"missing", ""}, {"wrong-scheme", "Basic abc"}, {"malformed-token", "Bearer not-a-jwt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/protected", nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			r.ServeHTTP(w, req)
			if w.Code != 401 {
				t.Fatalf("status=%d, want 401; body=%s", w.Code, w.Body.String())
			}
		})
	}
}
