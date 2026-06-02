package handler

import (
	"github.com/labstack/echo/v5"

	"github.com/yodzafar/url-shortener-app/internal/apperror"
	"github.com/yodzafar/url-shortener-app/internal/dto"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
	"github.com/yodzafar/url-shortener-app/internal/pkg/response"
	"github.com/yodzafar/url-shortener-app/internal/pkg/validation"
	"github.com/yodzafar/url-shortener-app/internal/usecase"
)

// AuthHandler exposes the authentication endpoints.
type AuthHandler struct {
	usecase   *usecase.AuthUsecase
	validator *validation.Validator
}

func NewAuthHandler(u *usecase.AuthUsecase, v *validation.Validator) *AuthHandler {
	return &AuthHandler{usecase: u, validator: v}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates an account and returns a JWT access/refresh token pair.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest	true	"Registration data"
//	@Success		201		{object}	dto.AuthResponse
//	@Failure		409		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	res, err := h.usecase.Register(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, res)
}

// Login godoc
//
//	@Summary		Log in
//	@Description	Verifies credentials and returns a JWT access/refresh token pair.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Login data"
//	@Success		200		{object}	dto.AuthResponse
//	@Failure		401		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	res, err := h.usecase.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return response.OK(c, res)
}

// Refresh godoc
//
//	@Summary		Refresh tokens
//	@Description	Exchanges a valid refresh token for a new access/refresh pair.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	dto.AuthResponse
//	@Failure		401		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	res, err := h.usecase.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return err
	}

	return response.OK(c, res)
}

// Me godoc
//
//	@Summary		Current user
//	@Description	Returns the authenticated user's profile.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	response.ErrorBody
//	@Router			/auth/me [get]
func (h *AuthHandler) Me(c *echo.Context) error {
	return response.OK(c, dto.NewUserResponse(appMiddleware.GetUser(c)))
}
