package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tobyrushton/padel-stats/libs/auth"
	"github.com/tobyrushton/padel-stats/libs/db/models"
)

type fakeAuthService struct {
	signupFn        func(context.Context, *auth.SignupInput) (*auth.AuthResult, error)
	signinFn        func(context.Context, *auth.SigninInput) (*auth.AuthResult, error)
	searchPlayersFn func(context.Context, string) (*auth.SearchPlayersResult, error)
	approveUserFn   func(context.Context, int64, int64) (*auth.User, error)
}

func (f *fakeAuthService) Signup(ctx context.Context, input *auth.SignupInput) (*auth.AuthResult, error) {
	if f.signupFn == nil {
		return nil, errors.New("signup function not configured")
	}
	return f.signupFn(ctx, input)
}

func (f *fakeAuthService) Signin(ctx context.Context, input *auth.SigninInput) (*auth.AuthResult, error) {
	if f.signinFn == nil {
		return nil, errors.New("signin function not configured")
	}
	return f.signinFn(ctx, input)
}

func (f *fakeAuthService) SearchPlayers(ctx context.Context, query string) (*auth.SearchPlayersResult, error) {
	if f.searchPlayersFn == nil {
		return nil, errors.New("search players function not configured")
	}
	return f.searchPlayersFn(ctx, query)
}

func (f *fakeAuthService) ApproveUser(ctx context.Context, adminUserID, userID int64) (*auth.User, error) {
	if f.approveUserFn == nil {
		return nil, errors.New("approve user function not configured")
	}

	return f.approveUserFn(ctx, adminUserID, userID)
}

func TestSignupSuccess(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		signupFn: func(ctx context.Context, input *auth.SignupInput) (*auth.AuthResult, error) {
			return &auth.AuthResult{
				User:  &auth.User{ID: 1, FirstName: "Jane", LastName: "Doe", Username: input.Username},
				Token: "jwt-token",
			}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBufferString(`{"firstName":"Jane","lastName":"Doe","username":"jane","password":"password123"}`))
	w := httptest.NewRecorder()

	h.Signup(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got auth.AuthResult
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, "jane", got.User.Username)
	assert.Equal(t, "jwt-token", got.Token)
}

func TestSigninSuccess(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		signinFn: func(ctx context.Context, input *auth.SigninInput) (*auth.AuthResult, error) {
			return &auth.AuthResult{
				User:  &auth.User{ID: 2, FirstName: "John", LastName: "Doe", Username: input.Username},
				Token: "signin-token",
			}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBufferString(`{"username":"john","password":"password123"}`))
	w := httptest.NewRecorder()

	h.Signin(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var got auth.AuthResult
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, "john", got.User.Username)
	assert.Equal(t, "signin-token", got.Token)
}

func TestSigninBadBody(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBufferString(`{bad json`))
	w := httptest.NewRecorder()

	h.Signin(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchPlayersSuccess(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		searchPlayersFn: func(ctx context.Context, query string) (*auth.SearchPlayersResult, error) {
			return &auth.SearchPlayersResult{
				Players: []*auth.User{{ID: 1, Username: "jane", FirstName: "Jane", LastName: "Doe"}},
			}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodGet, "/players/search?query=ja", nil)
	w := httptest.NewRecorder()

	h.SearchPlayers(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var got auth.SearchPlayersResult
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got.Players, 1)
	assert.Equal(t, "jane", got.Players[0].Username)
}

func TestSearchPlayersWithoutQueryReturnsDefaultList(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		searchPlayersFn: func(ctx context.Context, query string) (*auth.SearchPlayersResult, error) {
			assert.Equal(t, "", query)
			return &auth.SearchPlayersResult{
				Players: []*auth.User{{ID: 2, Username: "default-player"}},
			}, nil
		},
	}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodGet, "/players/search", nil)
	w := httptest.NewRecorder()

	h.SearchPlayers(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleAuthErrorMappings(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{err: auth.ErrUserExists, expected: http.StatusConflict},
		{err: auth.ErrUserNotFound, expected: http.StatusNotFound},
		{err: auth.ErrInvalidUsername, expected: http.StatusBadRequest},
		{err: auth.ErrInvalidPassword, expected: http.StatusUnauthorized},
		{err: auth.ErrUserPendingApproval, expected: http.StatusForbidden},
		{err: auth.ErrAdminAccessRequired, expected: http.StatusForbidden},
		{err: errors.New("boom"), expected: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		handleAuthError(w, tc.err)
		assert.Equal(t, tc.expected, w.Code)
	}
}

func TestApproveUserSuccess(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		approveUserFn: func(ctx context.Context, adminUserID, userID int64) (*auth.User, error) {
			assert.Equal(t, int64(10), adminUserID)
			assert.Equal(t, int64(20), userID)
			return &auth.User{ID: userID, Username: "pending", IsAcceptedByAdmin: true}, nil
		},
	}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			assert.Equal(t, "token-value", tokenString)
			return &models.Session{UserID: 10}, nil
		},
	})

	r := httptest.NewRequest(http.MethodPost, "/admin/users/20/approve", nil)
	r.Header.Set("Authorization", "Bearer token-value")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "20")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ApproveUser(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var got auth.User
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.True(t, got.IsAcceptedByAdmin)
}

func TestApproveUserUnauthorizedWithoutToken(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeSessionValidator{})

	r := httptest.NewRequest(http.MethodPost, "/admin/users/20/approve", nil)
	w := httptest.NewRecorder()

	h.ApproveUser(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApproveUserInvalidSession(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeSessionValidator{
		validateFn: func(ctx context.Context, tokenString string) (*models.Session, error) {
			return nil, errors.New("invalid")
		},
	})

	r := httptest.NewRequest(http.MethodPost, "/admin/users/20/approve", nil)
	r.Header.Set("Authorization", "Bearer token-value")
	w := httptest.NewRecorder()

	h.ApproveUser(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
