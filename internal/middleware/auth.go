package middleware

import (
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/yodzafar/url-shortener-app/internal/apperror"
	"github.com/yodzafar/url-shortener-app/internal/domain"
	"github.com/yodzafar/url-shortener-app/internal/pkg/token"
)

const UserContextKey = "user"

// AuthMiddleware authenticates requests using JWT Bearer access tokens.
type AuthMiddleware struct {
	tokens   *token.Manager
	userRepo domain.UserRepository
}

func NewAuthMiddleware(tokens *token.Manager, userRepo domain.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens, userRepo: userRepo}
}

// RequireAuth rejects requests without a valid access token, loading the user
// into the context on success.
func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw := bearerToken(c.Request().Header.Get("Authorization"))
			if raw == "" {
				return apperror.Unauthorized()
			}

			claims, err := m.tokens.Parse(raw)
			if err != nil || claims.Type != token.TokenAccess {
				return apperror.Unauthorized()
			}

			user, err := m.userRepo.FindByID(c.Request().Context(), claims.Subject)
			if err != nil {
				return apperror.Unauthorized()
			}

			c.Set(UserContextKey, user)

			return next(c)
		}
	}
}

// RequireRole authorizes the request only if the authenticated user has one of
// the allowed roles. It must run after RequireAuth (which loads the user).
func (m *AuthMiddleware) RequireRole(roles ...domain.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			user := GetUser(c)
			if user == nil {
				return apperror.Unauthorized()
			}

			for _, role := range roles {
				if user.Role == role {
					return next(c)
				}
			}

			return apperror.Forbidden()
		}
	}
}

// GetUser returns the authenticated user from the context, or nil.
func GetUser(c *echo.Context) *domain.User {
	u, _ := c.Get(UserContextKey).(*domain.User)
	return u
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}
