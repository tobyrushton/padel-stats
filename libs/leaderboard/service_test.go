package leaderboard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tobyrushton/padel-stats/libs/leaderboard"
)

type stubRepository struct {
	findSeasonLeaderboardFn  func(context.Context, int64) ([]*leaderboard.EntryRecord, error)
	findAllTimeLeaderboardFn func(context.Context) ([]*leaderboard.EntryRecord, error)
}

func (s *stubRepository) FindSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*leaderboard.EntryRecord, error) {
	if s.findSeasonLeaderboardFn == nil {
		return nil, errors.New("FindSeasonLeaderboard not configured")
	}

	return s.findSeasonLeaderboardFn(ctx, seasonID)
}

func (s *stubRepository) FindAllTimeLeaderboard(ctx context.Context) ([]*leaderboard.EntryRecord, error) {
	if s.findAllTimeLeaderboardFn == nil {
		return nil, errors.New("FindAllTimeLeaderboard not configured")
	}

	return s.findAllTimeLeaderboardFn(ctx)
}

func TestNewService(t *testing.T) {
	svc, err := leaderboard.NewService(nil)
	require.Error(t, err)
	assert.Nil(t, svc)
	assert.Equal(t, "leaderboard repository is required", err.Error())
}

func TestGetSeasonLeaderboard(t *testing.T) {
	svc, err := leaderboard.NewService(&stubRepository{
		findSeasonLeaderboardFn: func(ctx context.Context, seasonID int64) ([]*leaderboard.EntryRecord, error) {
			assert.Equal(t, int64(7), seasonID)
			return []*leaderboard.EntryRecord{
				{PlayerID: 2, Username: "b", ScoreDifference: 8, Wins: 4, Losses: 1, GamesPlayed: 5},
				{PlayerID: 1, Username: "a", ScoreDifference: 3, Wins: 3, Losses: 2, GamesPlayed: 5},
			}, nil
		},
		findAllTimeLeaderboardFn: func(ctx context.Context) ([]*leaderboard.EntryRecord, error) {
			return nil, nil
		},
	})
	require.NoError(t, err)

	result, err := svc.GetSeasonLeaderboard(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 2, result[1].Rank)
	assert.Equal(t, int64(2), result[0].PlayerID)
}

func TestGetSeasonLeaderboard_InvalidSeasonID(t *testing.T) {
	svc, err := leaderboard.NewService(&stubRepository{
		findSeasonLeaderboardFn: func(ctx context.Context, seasonID int64) ([]*leaderboard.EntryRecord, error) {
			return nil, nil
		},
		findAllTimeLeaderboardFn: func(ctx context.Context) ([]*leaderboard.EntryRecord, error) {
			return nil, nil
		},
	})
	require.NoError(t, err)

	result, err := svc.GetSeasonLeaderboard(context.Background(), 0)
	assert.ErrorIs(t, err, leaderboard.ErrInvalidSeasonID)
	assert.Nil(t, result)
}

func TestGetAllTimeLeaderboard(t *testing.T) {
	svc, err := leaderboard.NewService(&stubRepository{
		findSeasonLeaderboardFn: func(ctx context.Context, seasonID int64) ([]*leaderboard.EntryRecord, error) {
			return nil, nil
		},
		findAllTimeLeaderboardFn: func(ctx context.Context) ([]*leaderboard.EntryRecord, error) {
			return []*leaderboard.EntryRecord{{PlayerID: 9, Username: "u9", ScoreDifference: 12, Wins: 6, Losses: 1, GamesPlayed: 7}}, nil
		},
	})
	require.NoError(t, err)

	result, err := svc.GetAllTimeLeaderboard(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, int64(9), result[0].PlayerID)
}
