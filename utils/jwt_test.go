package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTAndValidateJWT(t *testing.T) {
	const userID int64 = 17
	const email = "user@example.com"

	token, err := GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("GenerateJWT() returned an error: %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %q, got %q", email, claims.Email)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.After(time.Now()) {
		t.Error("expected token to have a future expiration time")
	}
}

func TestValidateJWTRejectsInvalidTokens(t *testing.T) {
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: 1,
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	expiredTokenString, err := expiredToken.SignedString([]byte(jwtSecretKey))
	if err != nil {
		t.Fatalf("could not sign expired test token: %v", err)
	}

	wrongSignatureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: 2,
		Email:  "forged@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	wrongSignatureTokenString, err := wrongSignatureToken.SignedString([]byte("different-secret"))
	if err != nil {
		t.Fatalf("could not sign invalid test token: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{name: "malformed token", token: "not-a-jwt"},
		{name: "expired token", token: expiredTokenString},
		{name: "wrong signature", token: wrongSignatureTokenString},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if claims, err := ValidateJWT(test.token); err == nil {
				t.Fatalf("expected validation error, got claims %#v", claims)
			}
		})
	}
}
