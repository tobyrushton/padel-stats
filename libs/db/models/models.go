package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64     `bun:"id,pk,autoincrement"`
	FirstName      string    `bun:"first_name,notnull"`
	LastName       string    `bun:"last_name,notnull"`
	Username       string    `bun:"username,notnull,unique"`
	HashedPassword string    `bun:"hashed_password,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID        int64      `bun:"id,pk,autoincrement"`
	UserID    int64      `bun:"user_id,notnull"`
	User      *User      `bun:"rel:belongs-to,join:user_id=id"`
	TokenID   string     `bun:"token_id,notnull,unique"`
	ExpiresAt time.Time  `bun:"expires_at,notnull"`
	RevokedAt *time.Time `bun:"revoked_at"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp"`
}

// Season represents a padel season
type Season struct {
	bun.BaseModel `bun:"table:seasons,alias:s"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Year      int       `bun:"year,notnull"`
	StartDate time.Time `bun:"start_date,notnull"`
	EndDate   time.Time `bun:"end_date,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// Game represents a padel game (2v2)
type Game struct {
	bun.BaseModel `bun:"table:games,alias:g"`

	ID             int64     `bun:"id,pk,autoincrement"`
	SeasonID       int64     `bun:"season_id,notnull"`
	Season         *Season   `bun:"rel:belongs-to,join:season_id=id"`
	Team1Player1ID int64     `bun:"team1_player1_id,notnull"`
	Team1Player2ID int64     `bun:"team1_player2_id,notnull"`
	Team2Player1ID int64     `bun:"team2_player1_id,notnull"`
	Team2Player2ID int64     `bun:"team2_player2_id,notnull"`
	Team1Player1   *User     `bun:"rel:belongs-to,join:team1_player1_id=id"`
	Team1Player2   *User     `bun:"rel:belongs-to,join:team1_player2_id=id"`
	Team2Player1   *User     `bun:"rel:belongs-to,join:team2_player1_id=id"`
	Team2Player2   *User     `bun:"rel:belongs-to,join:team2_player2_id=id"`
	Team1Score     int       `bun:"team1_score,notnull"`
	Team2Score     int       `bun:"team2_score,notnull"`
	PlayedAt       time.Time `bun:"played_at,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
