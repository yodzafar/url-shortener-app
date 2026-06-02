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
		Where(sq.Eq{"email": email, "is_deleted": false}).
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
		Select("id", "first_name", "last_name", "middle_name", "gender", "birthdate", "email", "role", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"id": id, "is_deleted": false}).
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

func (r *userRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query, args, err := psql.
		Select("id", "first_name", "last_name", "middle_name", "gender", "birthdate", "email", "role", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"is_deleted": false}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}

	users := []*domain.User{}

	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (r *userRepo) Count(ctx context.Context) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("users").
		Where(sq.Eq{"is_deleted": false}).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var total int

	if err := r.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return total, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query, args, err := psql.
		Update("users").
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("middle_name", user.MiddleName).
		Set("gender", user.Gender).
		Set("birthdate", user.Birthdate).
		Set("is_active", user.IsActive).
		Where(sq.Eq{"id": user.ID, "is_deleted": false}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) SetRole(ctx context.Context, id string, role domain.Role) error {
	query, args, err := psql.
		Update("users").
		Set("role", string(role)).
		Where(sq.Eq{"id": id, "is_deleted": false}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build set-role query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("set user role: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	query, args, err := psql.
		Update("users").
		Set("is_deleted", true).
		Where(sq.Eq{"id": id, "is_deleted": false}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
