package helpers

import (
	"fmt"
	"log"
	"time"

	"math/rand"

	"github.com/golang-jwt/jwt/v4"
)

// Claims represents the custom claims embedded in the JWT (JSON Web Token).
//
// It extends the standard `jwt.RegisteredClaims` with an additional `UserID`
// field to uniquely identify the user associated with the token.
type Claims struct {
	jwt.RegisteredClaims        // RegisteredClaims: A set of standard JWT claims such as `iat`, `exp`, and `sub`.
	UserID               string // UserID: A custom claim representing the unique identifier of the user.
}

const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

// BuildJWTString generates a JWT (JSON Web Token) string containing a unique
// user ID and an expiration time.
//
// It uses the HS256 signing algorithm and embeds claims with a randomly generated
// user ID and an expiration time of 3 hours from the token's creation.
//
// Returns:
//   - string: The signed JWT string.
//   - error: An error if the token signing fails.
func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: GenerateRandomUserID(10),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserIDByToken extracts and validates the user ID from a JWT string.
//
// It parses the token using the HS256 signing algorithm and validates its
// signature and expiration time. If the token is valid, it extracts the
// "UserID" claim from the token's payload.
//
// Parameters:
//   - tokenString: The JWT string to parse and validate.
//
// Returns:
//   - string: The user ID extracted from the token if valid. Returns an empty
//     string if the token is invalid or parsing fails.
func GetUserIDByToken(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

	if err != nil {
		return ""
	}

	if !token.Valid {
		log.Println("Token is not valid")
		return ""
	}

	log.Println("Token is valid")
	return claims.UserID
}

// GenerateRandomUserID generates a random alphanumeric user ID of the specified length.
//
// It uses a seeded random number generator to produce a string consisting of
// uppercase, lowercase, and numeric characters.
//
// Parameters:
//   - length: The length of the user ID to generate.
//
// Returns:
//   - string: A randomly generated user ID.
func GenerateRandomUserID(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
