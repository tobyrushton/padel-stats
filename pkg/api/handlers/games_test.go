package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	gamedomain "github.com/tobyrushton/padel-stats/libs/games"
)

type fakeGamesService struct {
	createGameFn         func(context.Context, int64, *gamedomain.CreateGameInput) (*gamedomain.Game, error)
	listGamesForPlayerFn func(context.Context, int64) ([]*gamedomain.Game, error)
	getGameByIDFn        func(context.Context, int64) (*gamedomain.Game, error)
	deleteGameFn         func(context.Context, int64) error
}

func (f *fakeGamesService) CreateGame(ctx context.Context, creatorID int64, input *gamedomain.CreateGameInput) (*gamedomain.Game, error) {
	if f.createGameFn == nil {
		return nil, errors.New("create game function not configured")
	}
	return f.createGameFn(ctx, creatorID, input)
}

func (f *fakeGamesService) ListGamesForPlayer(ctx context.Context, playerID int64) ([]*gamedomain.Game, error) {
	if f.listGamesForPlayerFn == nil {
		return nil, errors.New("list games function not configured")
	}
	return f.listGamesForPlayerFn(ctx, playerID)
}

func (f *fakeGamesService) GetGameByID(ctx context.Context, gameID int64) (*gamedomain.Game, error) {
	if f.getGameByIDFn == nil {
		return nil, errors.New("get game function not configured")
	}
	return f.getGameByIDFn(ctx, gameID)
}

func (f *fakeGamesService) DeleteGame(ctx context.Context, gameID int64) error {
	if f.deleteGameFn == nil {
		return errors.New("delete game function not configured")
	}
	return f.deleteGameFn(ctx, gameID)
}

type fakeSessionValidator struct {
	validateFn func(context.Context, string) (*models.Session, error)
}

func (f *fakeSessionValidator) Validate(ctx context.Context, tokenString string) (*models.Session, error) {
	if f.validateFn == nil {
		return nil, errors.New("validate function not configured")
	}

	return f.validateFn(ctx, tokenString)
}

func TestCreateGameSuccess(t *testing.T) {
	playedAt := time.Now().UTC().Truncate(time.Second)
	h := NewGamesHandler(&fakeGamesService{
		createGameFn: func(ctx context.Context, creatorID int64, input *gamedomain.CreateGameInput) (*gamedomain.Game, error) {
			assert.Equal(t, int64(42), creatorID)
			return &gamedomain.Game{
				ID:             10,
				CreatorID:      creatorID,
				SeasonID:       12,
				Team1Player1ID: input.Team1Player1ID,
				Team1Player2ID: input.Team1Player2ID,
				Team2Player1ID: input.Team2Player1ID,
				Team2Player2ID: input.Team2Player2ID,
				Team1Score:     input.Team1Score,
				Team2Score:     input.Team2Score,
				PlayedAt:       playedAt,
			}, nil
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			assert.Equal(t, "token-value", tokenString)
			return &models.Session{UserID: 42}, nil
		},
	})

	body := `{"team1Player1Id":1,"team1Player2Id":2,"team2Player1Id":3,"team2Player2Id":4,"team1Score":6,"team2Score":4,"playedAt":"2026-03-30T12:00:00Z"}`
	r := httptest.NewRequest(http.MethodPost, "/games", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateGame(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got gamedomain.Game
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), got.ID)
	assert.Equal(t, int64(42), got.CreatorID)
}

func TestCreateGameBadBody(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42}, nil
		},
	})
	r := httptest.NewRequest(http.MethodPost, "/games", bytes.NewBufferString(`{bad json`))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateGame(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGameValidationError(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		createGameFn: func(ctx context.Context, creatorID int64, input *gamedomain.CreateGameInput) (*gamedomain.Game, error) {
			return nil, gamedomain.ErrNoSeasonForPlayedAt
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42}, nil
		},
	})

	body := `{"team1Player1Id":1,"team1Player2Id":2,"team2Player1Id":3,"team2Player2Id":4,"team1Score":6,"team2Score":4,"playedAt":"2026-03-30T12:00:00Z"}`
	r := httptest.NewRequest(http.MethodPost, "/games", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateGame(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGameUnauthorizedWithoutToken(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return nil, errors.New("should not be called")
		},
	})

	body := `{"team1Player1Id":1,"team1Player2Id":2,"team2Player1Id":3,"team2Player2Id":4,"team1Score":6,"team2Score":4,"playedAt":"2026-03-30T12:00:00Z"}`
	r := httptest.NewRequest(http.MethodPost, "/games", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateGame(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateGameRejectsUnknownSeasonIDField(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		createGameFn: func(ctx context.Context, creatorID int64, input *gamedomain.CreateGameInput) (*gamedomain.Game, error) {
			return nil, errors.New("should not be called")
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42}, nil
		},
	})

	body := `{"seasonId":999,"team1Player1Id":1,"team1Player2Id":2,"team2Player1Id":3,"team2Player2Id":4,"team1Score":6,"team2Score":4,"playedAt":"2026-03-30T12:00:00Z"}`
	r := httptest.NewRequest(http.MethodPost, "/games", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateGame(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGameByIDSuccess(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		getGameByIDFn: func(ctx context.Context, gameID int64) (*gamedomain.Game, error) {
			return &gamedomain.Game{ID: gameID, SeasonID: 1}, nil
		},
	}, &fakeSessionValidator{})

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/games/7", nil), "gameID", "7")
	w := httptest.NewRecorder()

	h.GetGameByID(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got gamedomain.Game
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), got.ID)
}

func TestGetGameByIDBadParam(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{}, &fakeSessionValidator{})
	r := withURLParam(httptest.NewRequest(http.MethodGet, "/games/x", nil), "gameID", "x")
	w := httptest.NewRecorder()

	h.GetGameByID(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGameByIDNotFound(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		getGameByIDFn: func(ctx context.Context, gameID int64) (*gamedomain.Game, error) {
			return nil, gamedomain.ErrGameNotFound
		},
	}, &fakeSessionValidator{})

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/games/9", nil), "gameID", "9")
	w := httptest.NewRecorder()

	h.GetGameByID(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListGamesForPlayerSuccess(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		listGamesForPlayerFn: func(ctx context.Context, playerID int64) ([]*gamedomain.Game, error) {
			return []*gamedomain.Game{{ID: 1}, {ID: 2}}, nil
		},
	}, &fakeSessionValidator{})

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/players/5/games", nil), "playerID", "5")
	w := httptest.NewRecorder()

	h.ListGamesForPlayer(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []*gamedomain.Game
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListGamesForPlayerBadParam(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{}, &fakeSessionValidator{})
	r := withURLParam(httptest.NewRequest(http.MethodGet, "/players/x/games", nil), "playerID", "x")
	w := httptest.NewRecorder()

	h.ListGamesForPlayer(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteGameSuccess(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		deleteGameFn: func(ctx context.Context, gameID int64) error {
			return nil
		},
	}, &fakeSessionValidator{})

	r := withURLParam(httptest.NewRequest(http.MethodDelete, "/games/3", nil), "gameID", "3")
	w := httptest.NewRecorder()

	h.DeleteGame(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestDeleteGameNotFound(t *testing.T) {
	h := NewGamesHandler(&fakeGamesService{
		deleteGameFn: func(ctx context.Context, gameID int64) error {
			return gamedomain.ErrGameNotFound
		},
	}, &fakeSessionValidator{})

	r := withURLParam(httptest.NewRequest(http.MethodDelete, "/games/3", nil), "gameID", "3")
	w := httptest.NewRecorder()

	h.DeleteGame(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGameErrorMappings(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{err: gamedomain.ErrGameNotFound, expected: http.StatusNotFound},
		{err: gamedomain.ErrNoSeasonForPlayedAt, expected: http.StatusBadRequest},
		{err: gamedomain.ErrSeasonOverlap, expected: http.StatusInternalServerError},
		{err: gamedomain.ErrInvalidPlayerID, expected: http.StatusBadRequest},
		{err: gamedomain.ErrDuplicatePlayers, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidScore, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidPlayedAt, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidCreatorID, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidGameID, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidDeleteGame, expected: http.StatusBadRequest},
		{err: gamedomain.ErrInvalidPlayerQuery, expected: http.StatusBadRequest},
		{err: errors.New("boom"), expected: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		handleGameError(w, tc.err)
		assert.Equal(t, tc.expected, w.Code)
	}
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
