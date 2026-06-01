// Package validation wraps go-playground/validator to produce localized,
// field-level validation errors shaped for the API error envelope.
package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	appi18n "github.com/yodzafar/url-shortener-app/i18n"
	"github.com/yodzafar/url-shortener-app/internal/apperror"
)

// Validator validates structs tagged with `validate:"..."` and reports errors
// keyed by their json field name with localized messages.
type Validator struct {
	v *validator.Validate
}

// New builds a Validator that reports fields by their json tag name.
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})

	return &Validator{v: v}
}

// Validate returns nil when s is valid, otherwise an *apperror.AppError (422)
// carrying localized, field-level details.
func (val *Validator) Validate(loc *goi18n.Localizer, s any) error {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// Non-validation error (e.g. invalid input type) — treat as bad request.
		return apperror.BadRequest().Wrap(err)
	}

	details := make(map[string][]string, len(verrs))
	for _, fe := range verrs {
		field := fe.Field() // json name, thanks to RegisterTagNameFunc
		msg := appi18n.T(loc, messageIDForTag(fe.Tag()), templateData(s, fe))
		details[field] = append(details[field], msg)
	}

	return apperror.Validation(details)
}

// messageIDForTag maps a validator tag to an i18n message key.
func messageIDForTag(tag string) string {
	switch tag {
	case "required":
		return "validation.required"
	case "email":
		return "validation.email"
	case "min":
		return "validation.min"
	case "max":
		return "validation.max"
	case "eqfield":
		return "validation.eqfield"
	default:
		return "validation.invalid"
	}
}

// templateData builds the i18n template data for a field error.
func templateData(s any, fe validator.FieldError) map[string]any {
	switch fe.Tag() {
	case "min", "max":
		return map[string]any{"Param": fe.Param()}
	case "eqfield":
		return map[string]any{"Field": jsonFieldName(s, fe.Param())}
	default:
		return nil
	}
}

// jsonFieldName resolves a Go struct field name to its json tag name on s.
// Falls back to a lowercased name when the field/tag is not found.
func jsonFieldName(s any, goField string) string {
	t := reflect.TypeOf(s)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		if f, ok := t.FieldByName(goField); ok {
			if name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]; name != "" && name != "-" {
				return name
			}
		}
	}

	return strings.ToLower(goField)
}
