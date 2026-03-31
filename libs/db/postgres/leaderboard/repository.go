package leaderboard

import (
	"context"
	"errors"

	leaderboarddomain "github.com/tobyrushton/padel-stats/libs/leaderboard"
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

func (r *Repository) FindSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*leaderboarddomain.EntryRecord, error) {
	query := `
SELECT
	ps.player_id,
	u.first_name,
	u.last_name,
	u.username,
	SUM(ps.score_diff) AS score_difference,
	SUM(CASE WHEN ps.score_diff > 0 THEN 1 ELSE 0 END) AS wins,
	SUM(CASE WHEN ps.score_diff < 0 THEN 1 ELSE 0 END) AS losses,
	COUNT(*) AS games_played
FROM (
	SELECT team1_player1_id AS player_id, (team1_score - team2_score) AS score_diff
	FROM games
	WHERE season_id = ?
	UNION ALL
	SELECT team1_player2_id AS player_id, (team1_score - team2_score) AS score_diff
	FROM games
	WHERE season_id = ?
	UNION ALL
	SELECT team2_player1_id AS player_id, (team2_score - team1_score) AS score_diff
	FROM games
	WHERE season_id = ?
	UNION ALL
	SELECT team2_player2_id AS player_id, (team2_score - team1_score) AS score_diff
	FROM games
	WHERE season_id = ?
) AS ps
JOIN users AS u ON u.id = ps.player_id
GROUP BY ps.player_id, u.first_name, u.last_name, u.username, u.id
ORDER BY score_difference DESC, wins DESC, u.username ASC, u.id ASC
`

	rows := make([]*leaderboarddomain.EntryRecord, 0)
	err := r.db.NewRaw(query, seasonID, seasonID, seasonID, seasonID).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *Repository) FindAllTimeLeaderboard(ctx context.Context) ([]*leaderboarddomain.EntryRecord, error) {
	query := `
SELECT
	ps.player_id,
	u.first_name,
	u.last_name,
	u.username,
	SUM(ps.score_diff) AS score_difference,
	SUM(CASE WHEN ps.score_diff > 0 THEN 1 ELSE 0 END) AS wins,
	SUM(CASE WHEN ps.score_diff < 0 THEN 1 ELSE 0 END) AS losses,
	COUNT(*) AS games_played
FROM (
	SELECT team1_player1_id AS player_id, (team1_score - team2_score) AS score_diff FROM games
	UNION ALL
	SELECT team1_player2_id AS player_id, (team1_score - team2_score) AS score_diff FROM games
	UNION ALL
	SELECT team2_player1_id AS player_id, (team2_score - team1_score) AS score_diff FROM games
	UNION ALL
	SELECT team2_player2_id AS player_id, (team2_score - team1_score) AS score_diff FROM games
) AS ps
JOIN users AS u ON u.id = ps.player_id
GROUP BY ps.player_id, u.first_name, u.last_name, u.username, u.id
ORDER BY score_difference DESC, wins DESC, u.username ASC, u.id ASC
`

	rows := make([]*leaderboarddomain.EntryRecord, 0)
	err := r.db.NewRaw(query).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
