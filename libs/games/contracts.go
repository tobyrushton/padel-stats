package games

import "time"

// CreateGameInput is the request payload for creating a game with four players.
type CreateGameInput struct {
	Team1Player1ID int64     `json:"team1Player1Id" format:"int64"`
	Team1Player2ID int64     `json:"team1Player2Id" format:"int64"`
	Team2Player1ID int64     `json:"team2Player1Id" format:"int64"`
	Team2Player2ID int64     `json:"team2Player2Id" format:"int64"`
	Team1Score     int       `json:"team1Score"`
	Team2Score     int       `json:"team2Score"`
	PlayedAt       time.Time `json:"playedAt" format:"date-time"`
}

// Player is an API-safe player contract for OpenAPI generation.
type Player struct {
	ID        int64  `json:"id" format:"int64"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
}

type Season struct {
	ID        int64     `json:"id" format:"int64"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate" format:"date-time"`
	EndDate   time.Time `json:"endDate" format:"date-time"`
}

// Game is an API-safe game contract.
type Game struct {
	ID             int64     `json:"id" format:"int64"`
	CreatorID      int64     `json:"creatorId" format:"int64"`
	SeasonID       int64     `json:"seasonId" format:"int64"`
	Team1Player1ID int64     `json:"team1Player1Id" format:"int64"`
	Team1Player2ID int64     `json:"team1Player2Id" format:"int64"`
	Team2Player1ID int64     `json:"team2Player1Id" format:"int64"`
	Team2Player2ID int64     `json:"team2Player2Id" format:"int64"`
	Team1Score     int       `json:"team1Score"`
	Team2Score     int       `json:"team2Score"`
	PlayedAt       time.Time `json:"playedAt" format:"date-time"`
	CreatedAt      time.Time `json:"createdAt" format:"date-time"`
	UpdatedAt      time.Time `json:"updatedAt" format:"date-time"`
	Team1Player1   *Player   `json:"team1Player1,omitempty"`
	Team1Player2   *Player   `json:"team1Player2,omitempty"`
	Team2Player1   *Player   `json:"team2Player1,omitempty"`
	Team2Player2   *Player   `json:"team2Player2,omitempty"`
	Season         *Season   `json:"season,omitempty"`
}

// PlayerRecord is a persistence-layer contract for user details attached to a game.
type PlayerRecord struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
}

type SeasonRecord struct {
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

// GameRecord is a persistence-layer contract independent from DB model structs.
type GameRecord struct {
	ID             int64
	CreatorID      int64
	SeasonID       int64
	Team1Player1ID int64
	Team1Player2ID int64
	Team2Player1ID int64
	Team2Player2ID int64
	Team1Score     int
	Team2Score     int
	PlayedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Team1Player1   *PlayerRecord
	Team1Player2   *PlayerRecord
	Team2Player1   *PlayerRecord
	Team2Player2   *PlayerRecord
	Season         *SeasonRecord
}

func gameFromRecord(record *GameRecord) *Game {
	if record == nil {
		return nil
	}

	return &Game{
		ID:             record.ID,
		CreatorID:      record.CreatorID,
		SeasonID:       record.SeasonID,
		Team1Player1ID: record.Team1Player1ID,
		Team1Player2ID: record.Team1Player2ID,
		Team2Player1ID: record.Team2Player1ID,
		Team2Player2ID: record.Team2Player2ID,
		Team1Score:     record.Team1Score,
		Team2Score:     record.Team2Score,
		PlayedAt:       record.PlayedAt,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		Team1Player1:   playerFromRecord(record.Team1Player1),
		Team1Player2:   playerFromRecord(record.Team1Player2),
		Team2Player1:   playerFromRecord(record.Team2Player1),
		Team2Player2:   playerFromRecord(record.Team2Player2),
		Season:         seasonFromRecord(record.Season),
	}
}

func playerFromRecord(record *PlayerRecord) *Player {
	if record == nil {
		return nil
	}

	return &Player{
		ID:        record.ID,
		FirstName: record.FirstName,
		LastName:  record.LastName,
		Username:  record.Username,
	}
}

func seasonFromRecord(record *SeasonRecord) *Season {
	if record == nil {
		return nil
	}

	return &Season{
		ID:        record.ID,
		Name:      record.Name,
		StartDate: record.StartDate,
		EndDate:   record.EndDate,
	}
}
