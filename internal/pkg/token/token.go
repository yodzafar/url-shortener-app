// Package token issues and validates stateless JWT access/refresh tokens.
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a token is malformed, expired or has an
// unexpected signing method.
var ErrInvalidToken = errors.New("invalid token")

// TokenType distinguishes access tokens from refresh tokens.
type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

// Claims is the JWT payload. Subject holds the user ID.
type Claims struct {
	Type TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Manager signs and verifies tokens with a shared HMAC secret.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager builds a token Manager.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// AccessTTL exposes the configured access-token lifetime.
func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

// GenerateAccessToken returns a signed access token and its expiry time.
func (m *Manager) GenerateAccessToken(userID string) (string, time.Time, error) {
	return m.generate(userID, TokenAccess, m.accessTTL)
}

// GenerateRefreshToken returns a signed refresh token.
func (m *Manager) GenerateRefreshToken(userID string) (string, error) {
	tok, _, err := m.generate(userID, TokenRefresh, m.refreshTTL)
	return tok, err
}

func (m *Manager) generate(userID string, typ TokenType, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)

	claims := Claims{
		Type: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, exp, nil
}

// Parse validates the token signature/expiry and returns its claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
