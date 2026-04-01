package games_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tobyrushton/padel-stats/libs/fakes"
	"github.com/tobyrushton/padel-stats/libs/games"
)

type ServiceTestSuite struct {
	suite.Suite
	repo    *fakes.FakeGamesRepository
	service *games.Service
	ctx     context.Context
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.repo = new(fakes.FakeGamesRepository)
	suite.ctx = context.Background()

	var err error
	suite.service, err = games.NewService(suite.repo)
	require.NoError(suite.T(), err)
}

func (suite *ServiceTestSuite) TestNewService_NilRepository() {
	service, err := games.NewService(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Equal(suite.T(), "games repository is required", err.Error())
}

func (suite *ServiceTestSuite) TestCreateGame_Success() {
	playedAt := time.Now().UTC()
	input := &games.CreateGameInput{
		SeasonID:       1,
		Team1Player1ID: 10,
		Team1Player2ID: 11,
		Team2Player1ID: 12,
		Team2Player2ID: 13,
		Team1Score:     6,
		Team2Score:     4,
		PlayedAt:       playedAt,
	}

	suite.repo.CreateGameStub = func(ctx context.Context, record *games.GameRecord) error {
		record.ID = 99
		record.CreatedAt = playedAt
		record.UpdatedAt = playedAt
		return nil
	}

	result, err := suite.service.CreateGame(suite.ctx, 77, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(99), result.ID)
	assert.Equal(suite.T(), int64(77), result.CreatorID)
	assert.Equal(suite.T(), int64(10), result.Team1Player1ID)
	assert.Equal(suite.T(), 1, suite.repo.CreateGameCallCount())
	_, createdRecord := suite.repo.CreateGameArgsForCall(0)
	assert.Equal(suite.T(), int64(77), createdRecord.CreatorID)
}

func (suite *ServiceTestSuite) TestCreateGame_DuplicatePlayers() {
	input := &games.CreateGameInput{
		SeasonID:       1,
		Team1Player1ID: 10,
		Team1Player2ID: 10,
		Team2Player1ID: 12,
		Team2Player2ID: 13,
		Team1Score:     6,
		Team2Score:     4,
		PlayedAt:       time.Now().UTC(),
	}

	result, err := suite.service.CreateGame(suite.ctx, 77, input)

	assert.ErrorIs(suite.T(), err, games.ErrDuplicatePlayers)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.repo.CreateGameCallCount())
}

func (suite *ServiceTestSuite) TestCreateGame_RepoError() {
	input := &games.CreateGameInput{
		SeasonID:       1,
		Team1Player1ID: 10,
		Team1Player2ID: 11,
		Team2Player1ID: 12,
		Team2Player2ID: 13,
		Team1Score:     6,
		Team2Score:     4,
		PlayedAt:       time.Now().UTC(),
	}

	suite.repo.CreateGameReturns(errors.New("insert failed"))

	result, err := suite.service.CreateGame(suite.ctx, 77, input)

	assert.EqualError(suite.T(), err, "insert failed")
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 1, suite.repo.CreateGameCallCount())
}

func (suite *ServiceTestSuite) TestCreateGame_InvalidCreatorID() {
	input := &games.CreateGameInput{
		SeasonID:       1,
		Team1Player1ID: 10,
		Team1Player2ID: 11,
		Team2Player1ID: 12,
		Team2Player2ID: 13,
		Team1Score:     6,
		Team2Score:     4,
		PlayedAt:       time.Now().UTC(),
	}

	result, err := suite.service.CreateGame(suite.ctx, 0, input)

	assert.ErrorIs(suite.T(), err, games.ErrInvalidCreatorID)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.repo.CreateGameCallCount())
}

func (suite *ServiceTestSuite) TestListGamesForPlayer_Success() {
	suite.repo.FindGamesByPlayerIDReturns([]*games.GameRecord{{ID: 1}, {ID: 2}}, nil)

	result, err := suite.service.ListGamesForPlayer(suite.ctx, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(1), result[0].ID)
	assert.Equal(suite.T(), 1, suite.repo.FindGamesByPlayerIDCallCount())
}

func (suite *ServiceTestSuite) TestListGamesForPlayer_InvalidPlayerID() {
	result, err := suite.service.ListGamesForPlayer(suite.ctx, 0)

	assert.ErrorIs(suite.T(), err, games.ErrInvalidPlayerQuery)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.repo.FindGamesByPlayerIDCallCount())
}

func (suite *ServiceTestSuite) TestGetGameByID_Success() {
	suite.repo.FindGameByIDReturns(&games.GameRecord{ID: 7}, nil)

	result, err := suite.service.GetGameByID(suite.ctx, 7)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(7), result.ID)
	assert.Equal(suite.T(), 1, suite.repo.FindGameByIDCallCount())
}

func (suite *ServiceTestSuite) TestGetGameByID_NotFound() {
	suite.repo.FindGameByIDReturns(nil, games.ErrGameNotFound)

	result, err := suite.service.GetGameByID(suite.ctx, 55)

	assert.ErrorIs(suite.T(), err, games.ErrGameNotFound)
	assert.Nil(suite.T(), result)
}

func (suite *ServiceTestSuite) TestDeleteGame_Success() {
	err := suite.service.DeleteGame(suite.ctx, 3)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, suite.repo.DeleteGameByIDCallCount())
	_, gameID := suite.repo.DeleteGameByIDArgsForCall(0)
	assert.Equal(suite.T(), int64(3), gameID)
}

func (suite *ServiceTestSuite) TestDeleteGame_InvalidID() {
	err := suite.service.DeleteGame(suite.ctx, 0)

	assert.ErrorIs(suite.T(), err, games.ErrInvalidDeleteGame)
	assert.Equal(suite.T(), 0, suite.repo.DeleteGameByIDCallCount())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
