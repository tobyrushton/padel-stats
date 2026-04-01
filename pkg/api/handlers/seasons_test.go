package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	seasonsdomain "github.com/tobyrushton/padel-stats/libs/seasons"
)

type fakeSeasonsService struct {
	getSeasonsFn      func(context.Context) ([]*seasonsdomain.Season, error)
	getActiveSeasonFn func(context.Context) (*seasonsdomain.Season, error)
	createSeasonFn    func(context.Context, *seasonsdomain.CreateSeasonInput) (*seasonsdomain.Season, error)
	endSeasonFn       func(context.Context, int64) (*seasonsdomain.Season, error)
}

func (f *fakeSeasonsService) GetSeasons(ctx context.Context) ([]*seasonsdomain.Season, error) {
	if f.getSeasonsFn == nil {
		return nil, errors.New("GetSeasons not configured")
	}

	return f.getSeasonsFn(ctx)
}

func (f *fakeSeasonsService) GetActiveSeason(ctx context.Context) (*seasonsdomain.Season, error) {
	if f.getActiveSeasonFn == nil {
		return nil, errors.New("GetActiveSeason not configured")
	}

	return f.getActiveSeasonFn(ctx)
}

func (f *fakeSeasonsService) CreateSeason(ctx context.Context, input *seasonsdomain.CreateSeasonInput) (*seasonsdomain.Season, error) {
	if f.createSeasonFn == nil {
		return nil, errors.New("CreateSeason not configured")
	}

	return f.createSeasonFn(ctx, input)
}

func (f *fakeSeasonsService) EndSeason(ctx context.Context, seasonID int64) (*seasonsdomain.Season, error) {
	if f.endSeasonFn == nil {
		return nil, errors.New("EndSeason not configured")
	}

	return f.endSeasonFn(ctx, seasonID)
}

func TestGetSeasonsSuccess(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getSeasonsFn: func(ctx context.Context) ([]*seasonsdomain.Season, error) {
			return []*seasonsdomain.Season{{ID: 1, Name: "S1"}, {ID: 2, Name: "S2"}}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodGet, "/seasons", nil)
	w := httptest.NewRecorder()

	h.GetSeasons(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got []*seasonsdomain.Season
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestGetSeasonsError(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getSeasonsFn: func(ctx context.Context) ([]*seasonsdomain.Season, error) {
			return nil, errors.New("boom")
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodGet, "/seasons", nil)
	w := httptest.NewRecorder()

	h.GetSeasons(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetActiveSeasonSuccess(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getActiveSeasonFn: func(ctx context.Context) (*seasonsdomain.Season, error) {
			return &seasonsdomain.Season{ID: 3, Name: "Current"}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodGet, "/seasons/active", nil)
	w := httptest.NewRecorder()

	h.GetActiveSeason(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var got seasonsdomain.Season
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), got.ID)
}

func TestCreateSeasonSuccess(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getActiveSeasonFn: func(ctx context.Context) (*seasonsdomain.Season, error) {
			return &seasonsdomain.Season{ID: 1, Name: "S1"}, nil
		},
		endSeasonFn: func(ctx context.Context, seasonID int64) (*seasonsdomain.Season, error) {
			assert.Equal(t, int64(1), seasonID)
			return &seasonsdomain.Season{ID: 1}, nil
		},
		createSeasonFn: func(ctx context.Context, input *seasonsdomain.CreateSeasonInput) (*seasonsdomain.Season, error) {
			assert.Equal(t, "S2", input.Name)
			assert.Equal(t, "2026-04-01", input.StartDate)
			return &seasonsdomain.Season{ID: 2, Name: "S2", StartDate: "2026-04-01"}, nil
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			assert.Equal(t, "token-value", tokenString)
			return &models.Session{UserID: 42, User: &models.User{IsAdmin: true}}, nil
		},
	})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)

	var got seasonsdomain.Season
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), got.ID)
	assert.Equal(t, "S2", got.Name)
}

func TestCreateSeasonUnauthorizedWithoutToken(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{}, &fakeSessionValidator{})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSeasonForbiddenForNonAdmin(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42, User: &models.User{IsAdmin: false}}, nil
		},
	})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateSeasonBadBody(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42, User: &models.User{IsAdmin: true}}, nil
		},
	})

	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(`{bad json`))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSeasonInvalidSession(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return nil, errors.New("invalid")
		},
	})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSeasonGetActiveSeasonError(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getActiveSeasonFn: func(ctx context.Context) (*seasonsdomain.Season, error) {
			return nil, errors.New("boom")
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42, User: &models.User{IsAdmin: true}}, nil
		},
	})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateSeasonEndSeasonError(t *testing.T) {
	h := NewSeasonsHandler(&fakeSeasonsService{
		getActiveSeasonFn: func(ctx context.Context) (*seasonsdomain.Season, error) {
			return &seasonsdomain.Season{ID: 1, Name: "S1"}, nil
		},
		endSeasonFn: func(ctx context.Context, seasonID int64) (*seasonsdomain.Season, error) {
			return nil, errors.New("boom")
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return &models.Session{UserID: 42, User: &models.User{IsAdmin: true}}, nil
		},
	})

	body := `{"name":"S2","startDate":"2026-04-01"}`
	r := httptest.NewRequest(http.MethodPost, "/seasons", bytes.NewBufferString(body))
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.CreateSeason(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSeasonsError(t *testing.T) {
	w := httptest.NewRecorder()
	handleSeasonsError(w, errors.New("boom"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
