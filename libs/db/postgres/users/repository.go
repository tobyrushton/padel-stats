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

func (r *Repository) CreateUser(ctx context.Context, user *auth.UserRecord) error {
	model := &models.User{
		ID:             user.ID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	_, err := r.db.NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return err
	}

	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return err
}

func (r *Repository) FindUserByUsername(ctx context.Context, username string) (*auth.UserRecord, error) {
	model := new(models.User)
	err := r.db.NewSelect().
		Model(model).
		Where("username = ?", username).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &auth.UserRecord{
		ID:             model.ID,
		FirstName:      model.FirstName,
		LastName:       model.LastName,
		Username:       model.Username,
		HashedPassword: model.HashedPassword,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}
