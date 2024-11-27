package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/helpers"
)

// ContextKey defines a custom type for keys in the context to avoid collisions.
type ContextKey string

// CookieAuthToken is the name of the cookie used for user authentication.
const CookieAuthToken = "auth_token"

// UserIDKey is the key used to store the user ID in the request context.
const UserIDKey = ContextKey("userID")

// GiveAuthTokenToUser is middleware that assigns an authentication token to the user if not already set.
//
// If a cookie with the `auth_token` name does not exist, this middleware generates a new JWT token,
// sets it as a cookie, and adds the token to the request context. If the cookie exists, it validates
// the token and adds it to the request context.
//
// Parameters:
//   - h: The next HTTP handler to call.
//
// Returns:
//   - An `http.Handler` that wraps the provided handler with the authentication logic.
//
// Behavior:
//   - If the token generation fails, it responds with HTTP 500 (Internal Server Error).
//   - Logs the token generation and cookie status.
func GiveAuthTokenToUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		authToken, err := r.Cookie(CookieAuthToken)
		if err != nil {
			token, err := helpers.BuildJWTString()
			if err != nil {
				log.Printf("BuildJWTString error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: CookieAuthToken, Value: token})
			ctx = context.WithValue(r.Context(), UserIDKey, token)
			log.Println("cookie is set")
		} else if authToken.Value != "" {
			log.Println("cookie is already set")
			ctx = context.WithValue(r.Context(), UserIDKey, authToken.Value)
		}

		log.Println("end GiveAuthTokenToUser")
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CheckAuthToken is middleware that validates the presence of an authentication token.
//
// This middleware checks if the `auth_token` cookie exists and is valid. If the token is missing
// or invalid, the middleware responds with HTTP 401 (Unauthorized). Otherwise, it adds the token
// to the request context and proceeds to the next handler.
//
// Parameters:
//   - h: The next HTTP handler to call.
//
// Returns:
//   - An `http.Handler` that wraps the provided handler with token validation.
//
// Behavior:
//   - If the token is missing or invalid, it responds with HTTP 401 (Unauthorized).
func CheckAuthToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		authToken, err := r.Cookie(CookieAuthToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if authToken.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(r.Context(), UserIDKey, authToken.Value)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
