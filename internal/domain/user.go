package domain

import (
	"context"
	"errors"
	"time"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// Role determines a user's access level.
type Role string

const (
	RoleAdmin Role = "admin" // full access to all resources
	RoleUser  Role = "user"  // access limited to own profile
)

// IsValid reports whether r is a known role.
func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)

type User struct {
	ID           string    `db:"id"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	MiddleName   string    `db:"middle_name"`
	Gender       Gender    `db:"gender"`
	Birthdate    string    `db:"birthdate"`
	Email        string    `db:"email"`
	Role         Role      `db:"role"`
	PasswordHash string    `db:"password_hash"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	IsDeleted    bool      `db:"is_deleted"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	SetRole(ctx context.Context, id string, role Role) error
}
