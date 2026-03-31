package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	leaderboarddomain "github.com/tobyrushton/padel-stats/libs/leaderboard"
)

type fakeLeaderboardService struct {
	getSeasonLeaderboardFn  func(context.Context, int64) ([]*leaderboarddomain.Entry, error)
	getAllTimeLeaderboardFn func(context.Context) ([]*leaderboarddomain.Entry, error)
}

func (f *fakeLeaderboardService) GetSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*leaderboarddomain.Entry, error) {
	if f.getSeasonLeaderboardFn == nil {
		return nil, errors.New("GetSeasonLeaderboard not configured")
	}

	return f.getSeasonLeaderboardFn(ctx, seasonID)
}

func (f *fakeLeaderboardService) GetAllTimeLeaderboard(ctx context.Context) ([]*leaderboarddomain.Entry, error) {
	if f.getAllTimeLeaderboardFn == nil {
		return nil, errors.New("GetAllTimeLeaderboard not configured")
	}

	return f.getAllTimeLeaderboardFn(ctx)
}

func TestGetSeasonLeaderboard(t *testing.T) {
	h := NewLeaderboardHandler(&fakeLeaderboardService{
		getSeasonLeaderboardFn: func(ctx context.Context, seasonID int64) ([]*leaderboarddomain.Entry, error) {
			assert.Equal(t, int64(3), seasonID)
			return []*leaderboarddomain.Entry{{Rank: 1, PlayerID: 1, Username: "alice", ScoreDifference: 10, Wins: 4, Losses: 1, GamesPlayed: 5}}, nil
		},
		getAllTimeLeaderboardFn: func(ctx context.Context) ([]*leaderboarddomain.Entry, error) {
			return nil, nil
		},
	})

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/seasons/3/leaderboard", nil), "seasonID", "3")
	w := httptest.NewRecorder()

	h.GetSeasonLeaderboard(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []*leaderboarddomain.Entry
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Rank)
}

func TestGetSeasonLeaderboard_InvalidParam(t *testing.T) {
	h := NewLeaderboardHandler(&fakeLeaderboardService{})
	r := withURLParam(httptest.NewRequest(http.MethodGet, "/seasons/x/leaderboard", nil), "seasonID", "x")
	w := httptest.NewRecorder()

	h.GetSeasonLeaderboard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllTimeLeaderboard(t *testing.T) {
	h := NewLeaderboardHandler(&fakeLeaderboardService{
		getSeasonLeaderboardFn: func(ctx context.Context, seasonID int64) ([]*leaderboarddomain.Entry, error) {
			return nil, nil
		},
		getAllTimeLeaderboardFn: func(ctx context.Context) ([]*leaderboarddomain.Entry, error) {
			return []*leaderboarddomain.Entry{{Rank: 1, PlayerID: 9, Username: "champ", ScoreDifference: 15, Wins: 8, Losses: 1, GamesPlayed: 9}}, nil
		},
	})

	r := httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	w := httptest.NewRecorder()

	h.GetAllTimeLeaderboard(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []*leaderboarddomain.Entry
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, int64(9), got[0].PlayerID)
}

func TestHandleLeaderboardError(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{err: leaderboarddomain.ErrInvalidSeasonID, expected: http.StatusBadRequest},
		{err: errors.New("boom"), expected: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		handleLeaderboardError(w, tc.err)
		assert.Equal(t, tc.expected, w.Code)
	}
}
