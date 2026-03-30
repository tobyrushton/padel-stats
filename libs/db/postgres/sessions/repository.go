package sessions

import (
	"context"
	"errors"
	"time"

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

func (r *Repository) Create(ctx context.Context, session *models.Session) error {
	_, err := r.db.NewInsert().Model(session).Exec(ctx)
	return err
}

func (r *Repository) FindByTokenID(ctx context.Context, tokenID string) (*models.Session, error) {
	session := new(models.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("token_id = ?", tokenID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *Repository) RevokeByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked_at = ?", revokedAt).
		Set("updated_at = ?", revokedAt).
		Where("token_id = ?", tokenID).
		Where("revoked_at IS NULL").
		Exec(ctx)
	return err
}
