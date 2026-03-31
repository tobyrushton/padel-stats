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
		IsAdmin:        user.IsAdmin,
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
		IsAdmin:        model.IsAdmin,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

func (r *Repository) SearchUsersByQuery(ctx context.Context, query string) ([]*auth.UserRecord, error) {
	modelsList := make([]models.User, 0)
	searchTerm := "%" + query + "%"

	err := r.db.NewSelect().
		Model(&modelsList).
		Where("username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", searchTerm, searchTerm, searchTerm).
		Order("username ASC").
		Limit(20).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*auth.UserRecord, 0, len(modelsList))
	for i := range modelsList {
		user := modelsList[i]
		result = append(result, &auth.UserRecord{
			ID:             user.ID,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Username:       user.Username,
			HashedPassword: user.HashedPassword,
			IsAdmin:        user.IsAdmin,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		})
	}

	return result, nil
}

func (r *Repository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	model := new(models.User)
	err := r.db.NewSelect().
		Model(model).
		Where("id = ? AND is_admin = true", userID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
