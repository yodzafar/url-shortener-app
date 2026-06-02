package usecase

import (
	"context"

	"github.com/yodzafar/url-shortener-app/internal/domain"
	"github.com/yodzafar/url-shortener-app/internal/dto"
)

// UserUsecase implements CRUD operations over users.
type UserUsecase struct {
	users domain.UserRepository
}

func NewUserUsecase(users domain.UserRepository) *UserUsecase {
	return &UserUsecase{users: users}
}

// List returns a page of users and the total count.
func (u *UserUsecase) List(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	offset := (page - 1) * pageSize

	users, err := u.users.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.users.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Get returns a single user by ID.
func (u *UserUsecase) Get(ctx context.Context, id string) (*domain.User, error) {
	return u.users.FindByID(ctx, id)
}

// Update replaces a user's profile fields and returns the updated user.
func (u *UserUsecase) Update(ctx context.Context, id string, req dto.UpdateUserRequest) (*domain.User, error) {
	user, err := u.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.MiddleName = req.MiddleName
	user.Gender = domain.Gender(req.Gender)
	user.Birthdate = req.Birthdate
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := u.users.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete soft-deletes a user.
func (u *UserUsecase) Delete(ctx context.Context, id string) error {
	return u.users.Delete(ctx, id)
}

// SetRole assigns a role to a user and returns the updated user.
func (u *UserUsecase) SetRole(ctx context.Context, id string, role domain.Role) (*domain.User, error) {
	if !role.IsValid() {
		return nil, domain.ErrForbidden
	}

	if err := u.users.SetRole(ctx, id, role); err != nil {
		return nil, err
	}

	return u.users.FindByID(ctx, id)
}
