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
	return r.GetSeasonByDate(ctx, time.Now().UTC())
}

func (r *Repository) GetSeasonByDate(ctx context.Context, playedAt time.Time) (*leaderboarddomain.Season, error) {
	lookupDate := playedAt.UTC().Format("2006-01-02")
	seasons := make([]*leaderboarddomain.Season, 0, 2)

	err := r.db.NewSelect().
		Model(&seasons).
		Where("?::date BETWEEN start_date::date AND end_date::date", lookupDate).
		OrderExpr("start_date DESC").
		Limit(2).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	switch len(seasons) {
	case 0:
		return nil, leaderboarddomain.ErrSeasonNotFoundForDate
	case 1:
		return seasons[0], nil
	default:
		return nil, leaderboarddomain.ErrMultipleSeasonsForDate
	}
}

func (r *Repository) CreateSeason(ctx context.Context, season *leaderboarddomain.CreateSeasonInput) (*leaderboarddomain.Season, error) {
	dbSeason := &leaderboarddomain.Season{
		Name:      season.Name,
		StartDate: season.StartDate,
		EndDate:   time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).String(),
	}

	_, err := r.db.NewInsert().Model(dbSeason).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return dbSeason, nil
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
