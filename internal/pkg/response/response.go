// Package response defines the standardized JSON envelope used by every API
// endpoint, both for success and error responses.
//
// Success: {"success": true,  "data": <obj|array>, "meta": {...}?, "error": null}
// Error:   {"success": false, "data": null,        "error": {"code","message","details"?}}
//
// Note: Data and Error intentionally have NO `omitempty` so they serialize as
// explicit JSON null. Callers with no payload must pass an untyped nil.
package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Envelope is the single response shape for the whole API.
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error"`
}

// ErrorBody describes a failure in a machine- and human-readable way.
type ErrorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}

// Meta carries response metadata such as pagination (only on list endpoints).
type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination is the standard list pagination block.
type Pagination struct {
	Page        int  `json:"page"`
	PageSize    int  `json:"page_size"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// OK writes a 200 success envelope.
func OK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 success envelope.
func Created(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// List writes a 200 success envelope with pagination metadata.
func List(c *echo.Context, data any, m *Meta) error {
	return c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: m})
}

// NewPagination builds a *Meta with computed total_pages/has_next/has_previous.
func NewPagination(page, pageSize, totalItems int) *Meta {
	if pageSize <= 0 {
		pageSize = 1
	}

	totalPages := (totalItems + pageSize - 1) / pageSize

	return &Meta{Pagination: &Pagination{
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}}
}
