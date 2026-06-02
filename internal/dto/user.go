package dto

import (
	"time"

	"github.com/yodzafar/url-shortener-app/internal/domain"
)

// UserResponse is the public representation of a user (never the domain model).
type UserResponse struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	MiddleName string    `json:"middleName"`
	Gender     string    `json:"gender"`
	Birthdate  string    `json:"birthdate"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
}

// NewUserResponse maps a domain user to its public DTO.
func NewUserResponse(u *domain.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Gender:     string(u.Gender),
		Birthdate:  u.Birthdate,
		Email:      u.Email,
		Role:       string(u.Role),
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
	}
}

// NewUserResponses maps a slice of domain users to their public DTOs.
func NewUserResponses(users []*domain.User) []*UserResponse {
	out := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, NewUserResponse(u))
	}
	return out
}

// UpdateUserRequest is the body for PUT /users/{id}. All fields are optional;
// provided values replace the stored profile.
type UpdateUserRequest struct {
	FirstName  string `json:"firstName" validate:"omitempty,max=100"`
	LastName   string `json:"lastName" validate:"omitempty,max=100"`
	MiddleName string `json:"middleName" validate:"omitempty,max=100"`
	Gender     string `json:"gender" validate:"omitempty,oneof=male female"`
	Birthdate  string `json:"birthdate" validate:"omitempty"`
	IsActive   *bool  `json:"isActive" validate:"omitempty"`
}

// UpdateRoleRequest is the body for PUT /users/{id}/role (admin only).
type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin user"`
}
