package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func Home(e *echo.Echo) {

	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "✓ Pong!")
	})
}
