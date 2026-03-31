package leaderboard

import (
	"context"
	"errors"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../fakes/leaderboard-repository.go . Repository
type Repository interface {
	FindSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*EntryRecord, error)
	FindAllTimeLeaderboard(ctx context.Context) ([]*EntryRecord, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("leaderboard repository is required")
	}

	return &Service{repo: repo}, nil
}

func (s *Service) GetSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*Entry, error) {
	if seasonID <= 0 {
		return nil, ErrInvalidSeasonID
	}

	records, err := s.repo.FindSeasonLeaderboard(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	return entriesFromRecords(records), nil
}

func (s *Service) GetAllTimeLeaderboard(ctx context.Context) ([]*Entry, error) {
	records, err := s.repo.FindAllTimeLeaderboard(ctx)
	if err != nil {
		return nil, err
	}

	return entriesFromRecords(records), nil
}

func entriesFromRecords(records []*EntryRecord) []*Entry {
	entries := make([]*Entry, len(records))
	for i, record := range records {
		entries[i] = &Entry{
			Rank:            i + 1,
			PlayerID:        record.PlayerID,
			FirstName:       record.FirstName,
			LastName:        record.LastName,
			Username:        record.Username,
			ScoreDifference: record.ScoreDifference,
			Wins:            record.Wins,
			Losses:          record.Losses,
			GamesPlayed:     record.GamesPlayed,
		}
	}

	return entries
}
