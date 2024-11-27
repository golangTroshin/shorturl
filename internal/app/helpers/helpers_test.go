package helpers_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golangTroshin/shorturl/internal/app/helpers"
	"github.com/stretchr/testify/assert"
)

func TestBuildJWTString(t *testing.T) {
	token, err := helpers.BuildJWTString()
	assert.NoError(t, err, "Building JWT should not return an error")
	assert.NotEmpty(t, token, "Generated JWT should not be empty")

	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3, "JWT should have 3 parts separated by '.'")
}

func TestGetUserIDByToken(t *testing.T) {
	token, err := helpers.BuildJWTString()
	assert.NoError(t, err, "Building JWT should not return an error")
	assert.NotEmpty(t, token, "Generated JWT should not be empty")

	userID := helpers.GetUserIDByToken(token)
	assert.NotEmpty(t, userID, "Extracted UserID should not be empty")
}

func TestGetUserIDByToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"
	userID := helpers.GetUserIDByToken(invalidToken)
	assert.Empty(t, userID, "Extracted UserID from an invalid token should be empty")
}

func TestGetUserIDByToken_ExpiredToken(t *testing.T) {
	expiredClaims := helpers.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		UserID: "testuser",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, err := token.SignedString([]byte("supersecretkey"))
	assert.NoError(t, err, "Signing expired token should not return an error")

	userID := helpers.GetUserIDByToken(tokenString)
	assert.Empty(t, userID, "Extracted UserID from an expired token should be empty")
}

func TestGenerateRandomUserID(t *testing.T) {
	length := 10
	randomID := helpers.GenerateRandomUserID(length)
	assert.Equal(t, length, len(randomID), "Generated UserID should have the correct length")

	for _, char := range randomID {
		assert.True(t, strings.ContainsAny(string(char), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"), "Generated UserID contains invalid characters")
	}
}
