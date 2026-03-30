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
	"github.com/tobyrushton/padel-stats/libs/auth"
)

type fakeAuthService struct {
	signupFn func(context.Context, *auth.SignupInput) (*auth.AuthResult, error)
	signinFn func(context.Context, *auth.SigninInput) (*auth.AuthResult, error)
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

func TestSignupSuccess(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{
		signupFn: func(ctx context.Context, input *auth.SignupInput) (*auth.AuthResult, error) {
			return &auth.AuthResult{
				User:  &auth.User{ID: 1, FirstName: "Jane", LastName: "Doe", Username: input.Username},
				Token: "jwt-token",
			}, nil
		},
	})

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
	})

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
	h := NewAuthHandler(&fakeAuthService{})

	r := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBufferString(`{bad json`))
	w := httptest.NewRecorder()

	h.Signin(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleAuthErrorMappings(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{err: auth.ErrUserExists, expected: http.StatusConflict},
		{err: auth.ErrInvalidUsername, expected: http.StatusBadRequest},
		{err: auth.ErrInvalidPassword, expected: http.StatusUnauthorized},
		{err: errors.New("boom"), expected: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		handleAuthError(w, tc.err)
		assert.Equal(t, tc.expected, w.Code)
	}
}
