package auth

import (
	"crypto/subtle"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Set by Init from the environment. There are no built-in fallbacks:
// the server refuses to start without them.
var JwtKey []byte
var AdminSecretCode string

const minJwtSecretLength = 32

// Init loads JWT_SECRET and ADMIN_SECRET from the environment.
// It returns an error if either one is missing or empty.
func Init() error {
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET is not set (add it to your .env file)")
	}
	if len(jwtSecret) < minJwtSecretLength {
		log.Printf("WARNING: JWT_SECRET is shorter than %d characters; use a longer random value", minJwtSecretLength)
	}

	adminSecret := strings.TrimSpace(os.Getenv("ADMIN_SECRET"))
	if adminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET is not set (add it to your .env file)")
	}

	JwtKey = []byte(jwtSecret)
	AdminSecretCode = adminSecret
	return nil
}

// IsAdminCode reports whether code matches the admin invite code.
// An empty code never matches.
func IsAdminCode(code string) bool {
	if code == "" || AdminSecretCode == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(code), []byte(AdminSecretCode)) == 1
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return JwtKey, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ExtractToken(authHeader string) string {
	return strings.TrimPrefix(authHeader, "Bearer ")
}
