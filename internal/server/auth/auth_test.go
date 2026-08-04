package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_GenerateJWTToken_rejects_overflowing_expiry(t *testing.T) {
	// Given
	const expiresSeconds = int(1<<63-1)/int(time.Second) + 1

	// When
	_, _, err := GenerateJWTToken(expiresSeconds)

	// Then
	if err == nil {
		t.Fatal("GenerateJWTToken() error = nil, want overflowing expiry error")
	}
}

func Test_GenerateJWTToken_rejects_expiry_below_minus_one(t *testing.T) {
	// Given
	const expiresSeconds = -2

	// When
	_, _, err := GenerateJWTToken(expiresSeconds)

	// Then
	if err == nil {
		t.Fatal("GenerateJWTToken() error = nil, want invalid expiry error")
	}
}

func Test_GenerateJWTToken_uses_seconds_for_expiry(t *testing.T) {
	// Given
	const expiresSeconds = 24 * 60 * 60
	before := time.Now()

	// When
	tokenString, _, err := GenerateJWTToken(expiresSeconds)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	claims := &jwt.RegisteredClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(tokenString, claims); err != nil {
		t.Fatalf("ParseUnverified() error = %v", err)
	}

	// Then
	if claims.ExpiresAt == nil {
		t.Fatal("GenerateJWTToken() did not set an expiry")
	}
	lifetime := claims.ExpiresAt.Time.Sub(before)
	if lifetime < 24*time.Hour-time.Second || lifetime > 24*time.Hour+time.Second {
		t.Fatalf("token lifetime = %v, want approximately 24h", lifetime)
	}
}
