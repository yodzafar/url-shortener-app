// Package usecase contains the application business logic, orchestrating
// repositories and infrastructure for the handlers.
package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yodzafar/url-shortener-app/internal/domain"
	"github.com/yodzafar/url-shortener-app/internal/dto"
	"github.com/yodzafar/url-shortener-app/internal/pkg/token"
)

// AuthUsecase implements registration, login and token refresh.
type AuthUsecase struct {
	users  domain.UserRepository
	tokens *token.Manager
}

func NewAuthUsecase(users domain.UserRepository, tokens *token.Manager) *AuthUsecase {
	return &AuthUsecase{users: users, tokens: tokens}
}

// Register creates a new user and issues a token pair.
func (u *AuthUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(hash),
	}

	if err := u.users.Create(ctx, user); err != nil {
		return nil, err // duplicate → domain.ErrUserAlreadyExists
	}

	return u.issueTokens(user)
}

// Login verifies credentials and issues a token pair.
func (u *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := u.users.FindByEmail(ctx, email)
	if err != nil {
		// Do not leak whether the email exists.
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredential
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredential
	}

	return u.issueTokens(user)
}

// Refresh validates a refresh token and issues a fresh token pair.
func (u *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	claims, err := u.tokens.Parse(refreshToken)
	if err != nil || claims.Type != token.TokenRefresh {
		return nil, domain.ErrUnauthorized
	}

	user, err := u.users.FindByID(ctx, claims.Subject)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return u.issueTokens(user)
}

// issueTokens builds the AuthResponse with a new access/refresh pair.
func (u *AuthUsecase) issueTokens(user *domain.User) (*dto.AuthResponse, error) {
	access, expiresAt, err := u.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refresh, err := u.tokens.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = "" // defensive; already json:"-"

	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
	}, nil
}
