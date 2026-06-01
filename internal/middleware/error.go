package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	appi18n "github.com/yodzafar/url-shortener-app/i18n"
	"github.com/yodzafar/url-shortener-app/internal/apperror"
	"github.com/yodzafar/url-shortener-app/internal/pkg/response"
)

// ErrorHandler is the centralized echo error handler. It converts any error
// into an *apperror.AppError and returns the standardized error envelope with
// its message translated according to the request's language (Accept-Language
// header or lang cookie).
type ErrorHandler struct {
	translator *appi18n.Translator
}

func NewErrorHandler(t *appi18n.Translator) *ErrorHandler {
	return &ErrorHandler{translator: t}
}

// Handle implements echo.HTTPErrorHandler.
func (h *ErrorHandler) Handle(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}

	appErr := h.toAppError(err)

	if appErr.Status >= http.StatusInternalServerError {
		log.Printf("error: %s %s -> %v", c.Request().Method, c.Request().URL.Path, err)
	}

	localizer := GetLocalizer(c)
	if localizer == nil {
		localizer = h.translator.NewLocalizer(c.Request())
	}

	body := response.Envelope{
		Success: false,
		Data:    nil,
		Error: &response.ErrorBody{
			Code:    appErr.Code,
			Message: appi18n.T(localizer, appErr.MessageID, appErr.Data),
			Details: appErr.Details,
		},
	}

	if writeErr := c.JSON(appErr.Status, body); writeErr != nil {
		log.Printf("error: writing response: %v", writeErr)
	}
}

// toAppError normalizes any error into a translatable AppError. Our own domain
// errors keep their specific message; echo's HTTP errors (404, 405, ...) are
// mapped to a code/message by their status code; everything else is a 500.
func (h *ErrorHandler) toAppError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// echo.StatusCode handles both *echo.HTTPError and echo's sentinel errors
	// (ErrNotFound, ErrMethodNotAllowed, ...); it returns 0 if none applies.
	code := echo.StatusCode(err)

	switch {
	case code == http.StatusNotFound:
		return apperror.New(code, apperror.CodeNotFound, apperror.MsgNotFound).Wrap(err)
	case code == http.StatusMethodNotAllowed:
		return apperror.New(code, apperror.CodeMethodNotAllowed, apperror.MsgMethodNotAllowed).Wrap(err)
	case code == http.StatusUnauthorized:
		return apperror.New(code, apperror.CodeUnauthorized, apperror.MsgUnauthorized).Wrap(err)
	case code == http.StatusBadRequest, code == http.StatusUnprocessableEntity:
		return apperror.New(http.StatusBadRequest, apperror.CodeBadRequest, apperror.MsgBadRequest).Wrap(err)
	case code >= http.StatusInternalServerError:
		return apperror.New(code, apperror.CodeInternal, apperror.MsgInternal).Wrap(err)
	case code != 0:
		return apperror.New(code, apperror.CodeBadRequest, apperror.MsgBadRequest).Wrap(err)
	default:
		// Not an echo HTTP error: fall back to domain-error mapping / 500.
		return apperror.From(err)
	}
}
