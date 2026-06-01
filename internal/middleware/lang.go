package middleware

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	appi18n "github.com/yodzafar/url-shortener-app/i18n"
)

const LocalizerKey = "localizer"

type LangMW struct {
	translator *appi18n.Translator
}

func NewLangMiddleware(t *appi18n.Translator) *LangMW {
	return &LangMW{translator: t}
}

func (m *LangMW) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		localizer := m.translator.NewLocalizer(c.Request())
		c.Set(LocalizerKey, localizer)
		return next(c)
	}
}

// RequestLogger logs each request through the app logger. On error it logs the
// localized message (matching what the client receives) instead of the raw
// i18n key, so logs read e.g. "Unauthorized" rather than "error.unauthorized".
func (m *LangMW) RequestLogger() echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(c *echo.Context, v echomw.RequestLoggerValues) error {
			logger := c.Logger()

			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Duration("latency", v.Latency),
				)
				return nil
			}

			localizer := GetLocalizer(c)
			if localizer == nil {
				localizer = m.translator.NewLocalizer(c.Request())
			}
			appErr := toAppError(v.Error)

			logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("error", appi18n.T(localizer, appErr.MessageID, appErr.Data)),
			)
			return nil
		},
	})
}

func GetLocalizer(c *echo.Context) *i18n.Localizer {
	l, _ := c.Get(LocalizerKey).(*i18n.Localizer)
	return l
}
