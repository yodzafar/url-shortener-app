package routes

import (
	"github.com/labstack/echo/v5"

	"github.com/yodzafar/url-shortener-app/internal/handler"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
)

func Auth(e *echo.Echo, h *handler.AuthHandler, authMW *appMiddleware.AuthMiddleware) {
	g := e.Group("/auth")

	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.GET("/me", h.Me, authMW.RequireAuth())
}
