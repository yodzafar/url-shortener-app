package domain

import (
	"context"
	"errors"
	"time"
)

type gender string

const (
	GenderMale   gender = "male"
	GenderFemale gender = "female"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrUnauthorized      = errors.New("unauthorized")
)

type User struct {
	ID           string    `json:"id" db:"id"`
	FirstName    string    `json:"firstName" db:"first_name"`
	LastName     string    `json:"lastName" db:"last_name"`
	MiddleName   string    `json:"middleName" db:"middle_name"`
	Gender       gender    `json:"gender" db:"gender"`
	Birthdate    string    `json:"birthdate" db:"birthdate"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	IsDeleted    bool      `json:"-" db:"is_deleted"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}
