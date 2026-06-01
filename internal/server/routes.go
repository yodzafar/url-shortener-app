package server

import (
	"github.com/labstack/echo/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
	"github.com/yodzafar/url-shortener-app/internal/server/routes"
)

func registerRoutes(e *echo.Echo, h Handlers, authMW *appMiddleware.AuthMiddleware) {
	routes.Home(e)
	routes.Auth(e, h.Auth, authMW)

	// Swagger UI: http://localhost:<port>/swagger/index.html
	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)))
}
