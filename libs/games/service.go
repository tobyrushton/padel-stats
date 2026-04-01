package games

import (
	"context"
	"errors"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../fakes/games-repository.go . GamesRepository
type GamesRepository interface {
	CreateGame(ctx context.Context, game *GameRecord) error
	FindGamesByPlayerID(ctx context.Context, playerID int64) ([]*GameRecord, error)
	FindGameByID(ctx context.Context, gameID int64) (*GameRecord, error)
	DeleteGameByID(ctx context.Context, gameID int64) error
}

type Service struct {
	repo GamesRepository
}

func NewService(repo GamesRepository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("games repository is required")
	}

	return &Service{repo: repo}, nil
}

func (s *Service) CreateGame(ctx context.Context, creatorID int64, input *CreateGameInput) (*Game, error) {
	if input == nil {
		return nil, ErrInvalidGameID
	}

	if creatorID <= 0 {
		return nil, ErrInvalidCreatorID
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	record := &GameRecord{
		CreatorID:      creatorID,
		SeasonID:       input.SeasonID,
		Team1Player1ID: input.Team1Player1ID,
		Team1Player2ID: input.Team1Player2ID,
		Team2Player1ID: input.Team2Player1ID,
		Team2Player2ID: input.Team2Player2ID,
		Team1Score:     input.Team1Score,
		Team2Score:     input.Team2Score,
		PlayedAt:       input.PlayedAt,
	}

	if err := s.repo.CreateGame(ctx, record); err != nil {
		return nil, err
	}

	return gameFromRecord(record), nil
}

func (s *Service) ListGamesForPlayer(ctx context.Context, playerID int64) ([]*Game, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerQuery
	}

	records, err := s.repo.FindGamesByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	games := make([]*Game, len(records))
	for i, record := range records {
		games[i] = gameFromRecord(record)
	}

	return games, nil
}

func (s *Service) GetGameByID(ctx context.Context, gameID int64) (*Game, error) {
	if gameID <= 0 {
		return nil, ErrInvalidGameID
	}

	record, err := s.repo.FindGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	return gameFromRecord(record), nil
}

func (s *Service) DeleteGame(ctx context.Context, gameID int64) error {
	if gameID <= 0 {
		return ErrInvalidDeleteGame
	}

	return s.repo.DeleteGameByID(ctx, gameID)
}
