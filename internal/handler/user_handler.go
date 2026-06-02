package handler

import (
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/yodzafar/url-shortener-app/internal/apperror"
	"github.com/yodzafar/url-shortener-app/internal/domain"
	"github.com/yodzafar/url-shortener-app/internal/dto"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
	"github.com/yodzafar/url-shortener-app/internal/pkg/response"
	"github.com/yodzafar/url-shortener-app/internal/pkg/validation"
	"github.com/yodzafar/url-shortener-app/internal/usecase"
)

// UserHandler exposes CRUD endpoints for users.
type UserHandler struct {
	usecase   *usecase.UserUsecase
	validator *validation.Validator
}

func NewUserHandler(u *usecase.UserUsecase, v *validation.Validator) *UserHandler {
	return &UserHandler{usecase: u, validator: v}
}

// List godoc
//
//	@Summary		List users
//	@Description	Returns a paginated list of users.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number (default 1)"
//	@Param			pageSize	query		int	false	"Page size (default 20, max 100)"
//	@Success		200			{array}		dto.UserResponse
//	@Failure		401			{object}	response.ErrorBody
//	@Router			/users [get]
func (h *UserHandler) List(c *echo.Context) error {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)

	if page < 1 {
		page = 1
	}
	switch {
	case pageSize < 1:
		pageSize = 1
	case pageSize > 100:
		pageSize = 100
	}

	users, total, err := h.usecase.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return err
	}

	return response.List(c, dto.NewUserResponses(users), response.NewPagination(page, pageSize, total))
}

// Get godoc
//
//	@Summary		Get a user
//	@Description	Returns a single user by ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	response.ErrorBody
//	@Failure		404	{object}	response.ErrorBody
//	@Router			/users/{id} [get]
func (h *UserHandler) Get(c *echo.Context) error {
	user, err := h.usecase.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return response.OK(c, dto.NewUserResponse(user))
}

// Update godoc
//
//	@Summary		Update a user
//	@Description	Updates a user's profile fields.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		dto.UpdateUserRequest	true	"Profile fields"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		401		{object}	response.ErrorBody
//	@Failure		404		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/users/{id} [put]
func (h *UserHandler) Update(c *echo.Context) error {
	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	user, err := h.usecase.Update(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return err
	}

	return response.OK(c, dto.NewUserResponse(user))
}

// UpdateMe godoc
//
//	@Summary		Update own profile
//	@Description	Updates the authenticated user's own profile.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.UpdateUserRequest	true	"Profile fields"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		401		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/users/me [put]
func (h *UserHandler) UpdateMe(c *echo.Context) error {
	me := appMiddleware.GetUser(c)
	if me == nil {
		return apperror.Unauthorized()
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	user, err := h.usecase.Update(c.Request().Context(), me.ID, req)
	if err != nil {
		return err
	}

	return response.OK(c, dto.NewUserResponse(user))
}

// SetRole godoc
//
//	@Summary		Set a user's role
//	@Description	Assigns a role (admin or user) to a user. Admin only.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		dto.UpdateRoleRequest	true	"Role"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		401		{object}	response.ErrorBody
//	@Failure		403		{object}	response.ErrorBody
//	@Failure		404		{object}	response.ErrorBody
//	@Failure		422		{object}	response.ErrorBody
//	@Router			/users/{id}/role [put]
func (h *UserHandler) SetRole(c *echo.Context) error {
	var req dto.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest().Wrap(err)
	}

	if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil {
		return err
	}

	user, err := h.usecase.SetRole(c.Request().Context(), c.Param("id"), domain.Role(req.Role))
	if err != nil {
		return err
	}

	return response.OK(c, dto.NewUserResponse(user))
}

// Delete godoc
//
//	@Summary		Delete a user
//	@Description	Soft-deletes a user by ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.ErrorBody
//	@Failure		404	{object}	response.ErrorBody
//	@Router			/users/{id} [delete]
func (h *UserHandler) Delete(c *echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return err
	}

	return response.OK(c, nil)
}

// queryInt reads an integer query parameter, falling back to def.
func queryInt(c *echo.Context, key string, def int) int {
	if v, err := strconv.Atoi(c.QueryParam(key)); err == nil {
		return v
	}
	return def
}
