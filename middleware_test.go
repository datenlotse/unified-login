package unified_login_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	unified_login "github.com/datenlotse/unified-login-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func generateJWT(secret string, userID string, scopes []string) string {
	claims := jwt.MapClaims{
		"sub":    userID,
		"scopes": scopes,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(secret))
	return signedToken
}

func TestCheckJWT(t *testing.T) {
	secret := "test_secret"
	middleware := unified_login.NewMiddleware(secret)
	userID := uuid.New().String()
	validToken := generateJWT(secret, userID, []string{"read", "write"})
	invalidToken := generateJWT("wrong_secret", userID, []string{"read"})

	tests := []struct {
		name       string
		authHeader string
		expectUser bool
	}{
		{"Valid token", "Bearer " + validToken, true},
		{"Invalid token", "Bearer " + invalidToken, false},
		{"No token", "", false},
		{"Malformed token", "Bearer123", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			handler := middleware.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user := r.Context().Value(unified_login.UserKey)
				if tc.expectUser && user == nil {
					t.Errorf("Expected user in context, got nil")
				}
				if !tc.expectUser && user != nil {
					t.Errorf("Expected nil user in context, got %v", user)
				}
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(w, req)
		})
	}
}

func TestMustBeAuthenticated(t *testing.T) {
	middleware := unified_login.NewMiddleware("test_secret")
	tests := []struct {
		name         string
		userContext  any
		expectedCode int
	}{
		{"Authenticated", unified_login.UserInformation{}, http.StatusOK},
		{"Unauthenticated", nil, http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = unified_login.SetUserInContext(req, tc.userContext)
			w := httptest.NewRecorder()
			handler := middleware.MustBeAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(w, req)
			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d", tc.expectedCode, w.Code)
			}
		})
	}
}

func TestMustHaveAnyScope(t *testing.T) {
	middleware := unified_login.NewMiddleware("test_secret")
	validUser := unified_login.UserInformation{UserId: uuid.New(), Scopes: []string{"read", "write"}}
	tests := []struct {
		name            string
		userContext     any
		requestedScopes []string
		expectedCode    int
	}{
		{"Has required scope", validUser, []string{"read"}, http.StatusOK},
		{"Missing required scope", validUser, []string{"admin"}, http.StatusForbidden},
		{"Unauthenticated", nil, []string{"read"}, http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = unified_login.SetUserInContext(req, tc.userContext)
			w := httptest.NewRecorder()
			handler := middleware.MustHaveAnyScope(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), tc.requestedScopes)
			handler.ServeHTTP(w, req)
			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d", tc.expectedCode, w.Code)
			}
		})
	}
}

func FuzzCheckJWT(f *testing.F) {
	secret := "test_secret"
	middleware := unified_login.NewMiddleware(secret)

	f.Fuzz(func(t *testing.T, token string) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler := middleware.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, req)
	})
}
