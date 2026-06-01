// Package apperror defines application-level errors that carry a machine code,
// an i18n message key and an HTTP status, so the central error handler can
// render a standardized, localized error envelope.
package apperror

import (
	"errors"
	"net/http"

	"github.com/yodzafar/url-shortener-app/internal/domain"
)

// AppError is an error that maps to an HTTP status and a translatable message.
type AppError struct {
	Status    int                 // HTTP status code
	Code      string              // machine-readable error code (e.g. VALIDATION_ERROR)
	MessageID string              // i18n message key
	Data      map[string]any      // optional template data for the translation
	Details   map[string][]string // optional field-level errors (validation)
	err       error               // wrapped underlying error (not exposed to clients)
}

func (e *AppError) Error() string {
	if e.err != nil {
		return e.MessageID + ": " + e.err.Error()
	}
	return e.MessageID
}

func (e *AppError) Unwrap() error { return e.err }

// New builds an AppError with the given status, machine code and message key.
func New(status int, code, messageID string) *AppError {
	return &AppError{Status: status, Code: code, MessageID: messageID}
}

// Wrap attaches an underlying error for logging while keeping the public message.
func (e *AppError) Wrap(err error) *AppError {
	clone := *e
	clone.err = err
	return &clone
}

// WithData attaches template data used when localizing the message.
func (e *AppError) WithData(data map[string]any) *AppError {
	clone := *e
	clone.Data = data
	return &clone
}

// WithDetails attaches field-level error messages (used for validation errors).
func (e *AppError) WithDetails(details map[string][]string) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

// Machine-readable error codes (the "code" field of the error envelope).
const (
	CodeInternal         = "INTERNAL"
	CodeNotFound         = "NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeBadRequest       = "BAD_REQUEST"
	CodeValidation       = "VALIDATION_ERROR"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeUserNotFound     = "USER_NOT_FOUND"
	CodeUserExists       = "USER_EXISTS"
	CodeInvalidCreds     = "INVALID_CREDENTIALS"
)

// Predefined i18n message keys (must exist in locales/*.json).
const (
	MsgInternal           = "error.internal"
	MsgNotFound           = "error.not_found"
	MsgMethodNotAllowed   = "error.method_not_allowed"
	MsgBadRequest         = "error.bad_request"
	MsgValidationFailed   = "error.validation_failed"
	MsgUnauthorized       = "error.unauthorized"
	MsgUserNotFound       = "error.user_not_found"
	MsgUserExists         = "error.user_exists"
	MsgInvalidCredentials = "error.invalid_credentials"
)

// Validation builds a 422 validation error carrying field-level details.
func Validation(details map[string][]string) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidation, MsgValidationFailed).WithDetails(details)
}

// Unauthorized builds a 401 error.
func Unauthorized() *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, MsgUnauthorized)
}

// BadRequest builds a 400 error (e.g. malformed JSON body).
func BadRequest() *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, MsgBadRequest)
}

// From maps a raw error to an *AppError. Known domain errors are translated to a
// specific status/code/message; anything else falls back to a generic 500.
func From(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return New(http.StatusNotFound, CodeUserNotFound, MsgUserNotFound).Wrap(err)
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return New(http.StatusConflict, CodeUserExists, MsgUserExists).Wrap(err)
	case errors.Is(err, domain.ErrInvalidCredential):
		return New(http.StatusUnauthorized, CodeInvalidCreds, MsgInvalidCredentials).Wrap(err)
	case errors.Is(err, domain.ErrUnauthorized):
		return New(http.StatusUnauthorized, CodeUnauthorized, MsgUnauthorized).Wrap(err)
	default:
		return New(http.StatusInternalServerError, CodeInternal, MsgInternal).Wrap(err)
	}
}
