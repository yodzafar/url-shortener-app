package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Lang = language.Tag

var (
	LangUZ = language.Uzbek
	LangRU = language.Russian
	LangEN = language.English

	DefaultLang = LangEN
)

type Translator struct {
	bundle  *i18n.Bundle
	matcher language.Matcher
}

func New(localesDir string) (*Translator, error) {
	bundle := i18n.NewBundle(DefaultLang)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	langs := []string{"uz", "ru", "en"}

	for _, lang := range langs {
		path := fmt.Sprintf("%s/%s.json", localesDir, lang)

		if _, err := bundle.LoadMessageFile(path); err != nil {
			return nil, fmt.Errorf("i18n: load %s: %w", path, err)
		}
	}

	// English first → the matcher falls back to English for unsupported locales.
	matcher := language.NewMatcher([]language.Tag{LangEN, LangRU, LangUZ})

	return &Translator{bundle: bundle, matcher: matcher}, nil
}

func (t *Translator) NewLocalizer(r *http.Request) *i18n.Localizer {
	lang := t.Detect(r)
	return i18n.NewLocalizer(t.bundle, lang.String())
}

func T(l *i18n.Localizer, messageID string, templateData ...map[string]any) string {
	cfg := &i18n.LocalizeConfig{MessageID: messageID}
	if len(templateData) > 0 && templateData[0] != nil {
		cfg.TemplateData = templateData[0]
	}

	msg, err := l.Localize(cfg)
	if err != nil {
		log.Printf("i18n: missing key %q: %v", messageID, err)
		return messageID
	}

	return msg
}

func (t *Translator) Detect(r *http.Request) Lang {
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		tags, _, err := language.ParseAcceptLanguage(acceptLang)
		if err == nil && len(tags) > 0 {
			matched, _, _ := t.matcher.Match(tags...)
			return matched
		}
	}

	return DefaultLang
}
