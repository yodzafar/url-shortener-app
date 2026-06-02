package routes

import (
	"github.com/labstack/echo/v5"

	"github.com/yodzafar/url-shortener-app/internal/domain"
	"github.com/yodzafar/url-shortener-app/internal/handler"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
)

func User(e *echo.Echo, h *handler.UserHandler, authMW *appMiddleware.AuthMiddleware) {
	g := e.Group("/users", authMW.RequireAuth())

	// Self-service: any authenticated user can update their own profile.
	g.PUT("/me", h.UpdateMe)

	// Admin only: full user management.
	admin := authMW.RequireRole(domain.RoleAdmin)
	g.GET("", h.List, admin)
	g.GET("/:id", h.Get, admin)
	g.PUT("/:id", h.Update, admin)
	g.DELETE("/:id", h.Delete, admin)
	g.PUT("/:id/role", h.SetRole, admin)
}
