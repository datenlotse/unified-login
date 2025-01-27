// Package unified_login provides integration with THDS unified login for Go projects
package unified_login

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// The key used by the middleware to store the user information
const UserKey = "THDS_UNIFIED_LOGIN_USER_ID"

type UserInformation struct {
	Scopes []string
	UserId uuid.UUID
}

type UnifiedLoginMiddleware struct {
	secret string
}

// Creates a new Middleware instance
//
// Provide the application secret that it used to sign the JWT
// This is identical to the one displayed in Unified Login Admin
func NewMiddleware(appSecret string) *UnifiedLoginMiddleware {
	m := &UnifiedLoginMiddleware{
		secret: appSecret,
	}

	return m
}

// This function checks for a presence of a valid JWT and sets the Userinformation inside the requests context
func (m UnifiedLoginMiddleware) CheckJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth_header := r.Header.Get("Authorization")
		if auth_header == "" {
			r = setUserInContext(r, nil)
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(auth_header, "Bearer") {
			r = setUserInContext(r, nil)
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(auth_header, " ")
		if len(parts) != 2 {
			r = setUserInContext(r, nil)
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(m.secret), nil
		})
		if err != nil {
			r = setUserInContext(r, nil)
			next.ServeHTTP(w, r)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userId, err := claims.GetSubject()
			if err != nil {
				r = setUserInContext(r, nil)
				next.ServeHTTP(w, r)
				return
			}

			scopes, ok := claims["scopes"].([]interface{})
			scopesStr := []string{}

			if ok {
				for _, v := range scopes {
					scopesStr = append(scopesStr, v.(string))
				}
			}

			user := UserInformation{
				UserId: uuid.MustParse(userId),
				Scopes: scopesStr,
			}
			fmt.Printf("User: %#v", user)

			r = setUserInContext(r, user)
		}

		next.ServeHTTP(w, r)
	})
}

// Checks if the user is authenticated
//
// If the user is authenticated it calls next, otherwise it will return a 401 error
func (m UnifiedLoginMiddleware) MustBeAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey)
		if user == nil {
			w.Header().Add("Content-Type", "plain/text")
			w.WriteHeader(401)
			w.Write([]byte("Token not provided or invalid"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Checks if the user has any of the provided scopes
//
// If the user has any of the scopes, it will call next, otherwise it will return a 403 error.
// This will return a 401 error if the user is not authenticated
func (m UnifiedLoginMiddleware) MustHaveAnyScope(next http.Handler, scopes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserKey).(UserInformation)
		if !ok {
			w.Header().Add("Content-Type", "plain/text")
			w.WriteHeader(401)
			w.Write([]byte("Token not provided or invalid"))
			return
		}

		for _, s := range scopes {
			for _, uS := range user.Scopes {
				if s == uS {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
	})
}

// Checks if the user has all of the provided scopes
//
// Will call next, if the user has all scopes, otherwise will return a 403 error.
// This will return a 401 error if the user is not authenticated
func (m UnifiedLoginMiddleware) MustHaveAllScopes(next http.Handler, scopes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserKey).(UserInformation)
		if !ok {
			w.Header().Add("Content-Type", "plain/text")
			w.WriteHeader(401)
			w.Write([]byte("Token not provided or invalid"))
			return
		}

		for _, s := range scopes {
			found := false
			for _, uS := range user.Scopes {
				if s == uS {
					found = true
					break
				}
			}
			if !found {
				w.Header().Add("Content-Type", "plain/text")
				w.WriteHeader(403)
				w.Write([]byte("Forbidden"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func setUserInContext(r *http.Request, user any) *http.Request {
	ctx := context.WithValue(r.Context(), UserKey, user)
	return r.Clone(ctx)
}
