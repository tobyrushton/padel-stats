package games

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tobyrushton/padel-stats/libs/db/models"
	gamedomain "github.com/tobyrushton/padel-stats/libs/games"
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

func (r *Repository) CreateGame(ctx context.Context, game *gamedomain.GameRecord) error {
	model := &models.Game{
		CreatorID:      game.CreatorID,
		SeasonID:       game.SeasonID,
		Team1Player1ID: game.Team1Player1ID,
		Team1Player2ID: game.Team1Player2ID,
		Team2Player1ID: game.Team2Player1ID,
		Team2Player2ID: game.Team2Player2ID,
		Team1Score:     game.Team1Score,
		Team2Score:     game.Team2Score,
		PlayedAt:       game.PlayedAt,
	}

	_, err := r.db.NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return err
	}

	game.ID = model.ID
	game.CreatedAt = model.CreatedAt
	game.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *Repository) FindGamesByPlayerID(ctx context.Context, playerID int64) ([]*gamedomain.GameRecord, error) {
	var gameModels []models.Game
	err := r.db.NewSelect().
		Model(&gameModels).
		Where("team1_player1_id = ? OR team1_player2_id = ? OR team2_player1_id = ? OR team2_player2_id = ?", playerID, playerID, playerID, playerID).
		Relation("Team1Player1").
		Relation("Team1Player2").
		Relation("Team2Player1").
		Relation("Team2Player2").
		OrderExpr("played_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	games := make([]*gamedomain.GameRecord, len(gameModels))
	for i := range gameModels {
		games[i] = gameRecordFromModel(&gameModels[i])
	}

	return games, nil
}

func (r *Repository) FindGameByID(ctx context.Context, gameID int64) (*gamedomain.GameRecord, error) {
	model := new(models.Game)
	err := r.db.NewSelect().
		Model(model).
		Relation("Team1Player1").
		Relation("Team1Player2").
		Relation("Team2Player1").
		Relation("Team2Player2").
		Where("g.id = ?", gameID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gamedomain.ErrGameNotFound
		}
		return nil, err
	}

	return gameRecordFromModel(model), nil
}

func (r *Repository) DeleteGameByID(ctx context.Context, gameID int64) error {
	result, err := r.db.NewDelete().
		Model((*models.Game)(nil)).
		Where("id = ?", gameID).
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return gamedomain.ErrGameNotFound
	}

	return nil
}

func gameRecordFromModel(model *models.Game) *gamedomain.GameRecord {
	if model == nil {
		return nil
	}

	return &gamedomain.GameRecord{
		ID:             model.ID,
		CreatorID:      model.CreatorID,
		SeasonID:       model.SeasonID,
		Team1Player1ID: model.Team1Player1ID,
		Team1Player2ID: model.Team1Player2ID,
		Team2Player1ID: model.Team2Player1ID,
		Team2Player2ID: model.Team2Player2ID,
		Team1Score:     model.Team1Score,
		Team2Score:     model.Team2Score,
		PlayedAt:       model.PlayedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		Team1Player1:   playerRecordFromModel(model.Team1Player1),
		Team1Player2:   playerRecordFromModel(model.Team1Player2),
		Team2Player1:   playerRecordFromModel(model.Team2Player1),
		Team2Player2:   playerRecordFromModel(model.Team2Player2),
	}
}

func playerRecordFromModel(model *models.User) *gamedomain.PlayerRecord {
	if model == nil {
		return nil
	}

	return &gamedomain.PlayerRecord{
		ID:        model.ID,
		FirstName: model.FirstName,
		LastName:  model.LastName,
		Username:  model.Username,
	}
}
