package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axywe/filmotheka_vk/internal/middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func generateToken(role float64, secretKey string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": role,
	})
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

func TestRoleCheckMiddleware(t *testing.T) {
	secretKey := "your_secret_key" 

	tests := []struct {
		name           string
		tokenRole      float64
		requestMethod  string
		expectedStatus int
	}{
		{"ValidTokenRole1", 1, http.MethodGet, http.StatusOK},
		{"ValidTokenRole2Get", 2, http.MethodGet, http.StatusOK},
		{"ValidTokenRole2Post", 2, http.MethodPost, http.StatusForbidden},
		{"NoToken", 0, http.MethodGet, http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(test.requestMethod, "/", bytes.NewBufferString(""))
			if test.name != "NoToken" {
				token := generateToken(test.tokenRole, secretKey)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.RoleCheckMiddleware(handler).ServeHTTP(rr, req)

			assert.Equal(t, test.expectedStatus, rr.Code)
		})
	}
}

func TestRoleCheckMiddlewareWithError(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"InvalidToken", "invalid.token.here", http.StatusUnauthorized},
		{"EmptyToken", "", http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+test.token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			middleware.RoleCheckMiddleware(handler).ServeHTTP(rr, req)

			assert.Equal(t, test.expectedStatus, rr.Code)
		})
	}
}

func TestRoleCheckMiddlewareWithInvalidClaims(t *testing.T) {
	secretKey := "your_secret_key"

	generateInvalidClaimsToken := func() string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"not_role": "should_fail",
		})
		tokenString, _ := token.SignedString([]byte(secretKey))
		return tokenString
	}

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	token := generateInvalidClaimsToken()
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	middleware.RoleCheckMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
