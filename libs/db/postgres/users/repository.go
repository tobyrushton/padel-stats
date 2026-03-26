package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tobyrushton/padel-stats/libs/auth"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	return &Repository{db: db}, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *Repository) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("username = ?", username).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
