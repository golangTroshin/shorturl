package helpers

import (
	"bytes"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"math/big"
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

// GetTLSCertificate generates a self-signed TLS certificate and returns
// its PEM-encoded certificate and private key.
//
// It creates a certificate with the following properties:
// - Serial number: 7657
// - Organization: "Vladimir Troshin"
// - Country: "Cohort 33"
// - IP addresses: IPv4 127.0.0.1, IPv6 loopback address
// - Validity: 10 years
// - Subject key ID: [1, 2, 3, 4, 6]
// - ExtKeyUsage: ClientAuth and ServerAuth
// - KeyUsage: DigitalSignature
//
// It also generates an RSA private key with a 4096-bit length.
//
// The function returns two byte slices: the PEM-encoded certificate
// and the PEM-encoded private key.
//
// Example usage:
//
//	certPEM, keyPEM := GetTLSCertificate()
func GetTLSCertificate() ([]byte, []byte) {
	// Create a certificate template
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(7657),
		Subject: pkix.Name{
			Organization: []string{"Vladimir Troshin"},
			Country:      []string{"Cohort 33"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), // 10 years validity
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(cryptoRand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(cryptoRand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Encode the certificate and key in PEM format
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM.Bytes(), privateKeyPEM.Bytes()
}

// SaveToFile writes the given content to a file with the specified filename.
//
// It creates the file if it does not exist, and writes the content to it.
// If the file creation or writing fails, it logs the error and terminates the program.
//
// Example usage:
//
//	SaveToFile("cert.pem", certPEM)
//	SaveToFile("key.pem", keyPEM)
func SaveToFile(filename string, content []byte) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		log.Fatalf("Failed to write to file %s: %v", filename, err)
	}
}
