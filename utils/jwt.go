// Package utils provides shared helpers for authentication, password handling,
// logging, and error reporting.
package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecretKey = "your-secret-key"

// JWTClaims contains the application-specific claims embedded in an
// authentication token.
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed HS256 token for the specified user. The token
// expires two hours after it is generated.
func GenerateJWT(userID int64, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		},
	})
	return token.SignedString([]byte(jwtSecretKey))
}

// ValidateJWT parses and validates a signed authentication token, including
// its signing algorithm and expiration, and returns its claims.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
