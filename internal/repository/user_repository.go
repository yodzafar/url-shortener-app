package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yodzafar/url-shortener-app/internal/domain"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	user.IsActive = true

	query, args, err := psql.
		Insert("users").
		Columns("email", "password_hash", "is_active").
		Values(user.Email, user.PasswordHash, user.IsActive).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return domain.ErrUserAlreadyExists
		}

		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := psql.
		Select("id", "email", "password_hash", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	user := &domain.User{}

	if err := r.db.GetContext(ctx, user, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query, args, err := psql.
		Select("id", "first_name", "last_name", "middle_name", "gender", "birthdate", "email", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select query %w", err)
	}

	user := &domain.User{}

	if err := r.db.GetContext(ctx, user, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}
