package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64
	FirstName      string
	LastName       string
	Username       string `bun:"username,unique"`
	HashedPassword string
	CreatedAt      time.Time `bun:"created_at,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,default:current_timestamp"`
}

// Season represents a padel season
type Season struct {
	bun.BaseModel `bun:"table:seasons,alias:s"`

	ID        int64
	Name      string
	Year      int
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp"`
}

// Game represents a padel game (2v2)
type Game struct {
	bun.BaseModel `bun:"table:games,alias:g"`

	ID             int64
	SeasonID       int64
	Season         *Season `bun:"rel:has-one,join:season_id=id"`
	Team1Player1ID int64
	Team1Player2ID int64
	Team2Player1ID int64
	Team2Player2ID int64
	Team1Player1   *User `bun:"rel:has-one,join:team1_player1_id=id"`
	Team1Player2   *User `bun:"rel:has-one,join:team1_player2_id=id"`
	Team2Player1   *User `bun:"rel:has-one,join:team2_player1_id=id"`
	Team2Player2   *User `bun:"rel:has-one,join:team2_player2_id=id"`
	Team1Score     int
	Team2Score     int
	PlayedAt       time.Time
	CreatedAt      time.Time `bun:"created_at,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,default:current_timestamp"`
}
