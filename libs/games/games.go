package games

import (
	"errors"
)

var (
	ErrGameNotFound        = errors.New("game not found")
	ErrInvalidPlayerID     = errors.New("invalid player id")
	ErrDuplicatePlayers    = errors.New("players in a game must be unique")
	ErrInvalidScore        = errors.New("scores must be greater than or equal to 0")
	ErrInvalidPlayedAt     = errors.New("played at is required")
	ErrNoSeasonForPlayedAt = errors.New("no season found for played at")
	ErrSeasonOverlap       = errors.New("multiple seasons found for played at")
	ErrInvalidGameID       = errors.New("invalid game id")
	ErrInvalidDeleteGame   = errors.New("invalid game id for delete")
	ErrInvalidPlayerQuery  = errors.New("invalid player id for query")
	ErrInvalidCreatorID    = errors.New("invalid creator id")
)

func (in *CreateGameInput) Validate() error {
	playerIDs := []int64{in.Team1Player1ID, in.Team1Player2ID, in.Team2Player1ID, in.Team2Player2ID}
	seen := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if playerID <= 0 {
			return ErrInvalidPlayerID
		}
		if _, exists := seen[playerID]; exists {
			return ErrDuplicatePlayers
		}
		seen[playerID] = struct{}{}
	}

	if in.Team1Score < 0 || in.Team2Score < 0 {
		return ErrInvalidScore
	}

	if in.PlayedAt.IsZero() {
		return ErrInvalidPlayedAt
	}

	return nil
}
