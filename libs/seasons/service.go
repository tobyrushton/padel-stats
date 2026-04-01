package seasons

import (
	"context"
	"errors"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../fakes/seasons-repository.go . SeasonsRepository
type SeasonsRepository interface {
	GetSeasons(ctx context.Context) ([]*Season, error)
	GetActiveSeason(ctx context.Context) (*Season, error)
	CreateSeason(ctx context.Context, season *Season) (*Season, error)
	EndSeason(ctx context.Context, seasonID int64) (*Season, error)
}

type Service struct {
	repo SeasonsRepository
}

func NewService(repo SeasonsRepository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("seasons repository is required")
	}

	return &Service{repo: repo}, nil
}

func (s *Service) ListSeasons(ctx context.Context) ([]*Season, error) {
	return s.repo.GetSeasons(ctx)
}

func (s *Service) GetActiveSeason(ctx context.Context) (*Season, error) {
	return s.repo.GetActiveSeason(ctx)
}

func (s *Service) CreateSeason(ctx context.Context, season *Season) (*Season, error) {
	if season == nil {
		return nil, errors.New("season is required")
	}

	return s.repo.CreateSeason(ctx, season)
}

func (s *Service) EndSeason(ctx context.Context, seasonID int64) (*Season, error) {
	if seasonID <= 0 {
		return nil, errors.New("invalid season id")
	}

	return s.repo.EndSeason(ctx, seasonID)
}
