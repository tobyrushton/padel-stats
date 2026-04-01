package seasons

import (
	"context"
	"errors"
	"time"

	leaderboarddomain "github.com/tobyrushton/padel-stats/libs/seasons"
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

func (r *Repository) GetSeasons(ctx context.Context) ([]*leaderboarddomain.Season, error) {
	seasons := make([]*leaderboarddomain.Season, 0)
	err := r.db.NewSelect().Model(&seasons).OrderBy("start_date", bun.OrderAsc).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return seasons, nil
}

func (r *Repository) GetActiveSeason(ctx context.Context) (*leaderboarddomain.Season, error) {
	season := &leaderboarddomain.Season{}
	err := r.db.NewSelect().
		Model(season).
		Where("? BETWEEN start_date AND end_date", time.Now().UTC().String()).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return season, nil
}

func (r *Repository) CreateSeason(ctx context.Context, season *leaderboarddomain.Season) (*leaderboarddomain.Season, error) {
	_, err := r.db.NewInsert().Model(season).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return season, nil
}

func (r *Repository) EndSeason(ctx context.Context, seasonID int64) (*leaderboarddomain.Season, error) {
	season := &leaderboarddomain.Season{}
	err := r.db.NewSelect().Model(season).Where("id = ?", seasonID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	season.EndDate = time.Now().UTC().String()

	_, err = r.db.NewUpdate().Model(season).Where("id = ?", seasonID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return season, nil
}
