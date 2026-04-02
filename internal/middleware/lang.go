package middleware

import (
	"github.com/labstack/echo/v5"
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

func GetLocalizer(c *echo.Context) *i18n.Localizer {
	l, _ := c.Get(LocalizerKey).(*i18n.Localizer)
	return l
}
