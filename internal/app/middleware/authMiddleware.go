package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/helpers"
)

type ContextKey string

const CookieAuthToken = "auth_token"

const UserIDKey = ContextKey("userID")

func GiveAuthTokenToUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			r.Header.Set("Authorization", token)
			ctx = context.WithValue(r.Context(), UserIDKey, token)
			log.Println("cookie is set")
		} else if authToken.Value != "" {
			log.Println("cookie is already set")
			ctx = context.WithValue(r.Context(), UserIDKey, authToken.Value)
		}

		log.Println("end GiveAuthTokenToUser")
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}
