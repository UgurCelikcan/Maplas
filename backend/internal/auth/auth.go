package auth

import (
	"fmt"
	"strings"

	"backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte(utils.GetEnv("JWT_SECRET", "my_super_secret_key_2026"))
var AdminSecretCode = utils.GetEnv("ADMIN_SECRET", "Maplas-2026")

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return JwtKey, nil })
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
