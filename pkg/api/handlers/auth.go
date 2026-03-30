package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tobyrushton/padel-stats/libs/auth"
)

type AuthService interface {
	Signup(ctx context.Context, input *auth.SignupInput) (*auth.AuthResult, error)
	Signin(ctx context.Context, input *auth.SigninInput) (*auth.AuthResult, error)
}

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.Signup)
		r.Post("/signin", h.Signin)
	})
}

// Signup registers a new user and returns an auth token.
// @Summary Sign up
// @Description Register a new user account.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.SignupInput true "Signup payload"
// @Success 201 {object} auth.AuthResult
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var input auth.SignupInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.authService.Signup(r.Context(), &input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// Signin authenticates an existing user and returns an auth token.
// @Summary Sign in
// @Description Authenticate with username and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.SigninInput true "Signin payload"
// @Success 200 {object} auth.AuthResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/signin [post]
func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var input auth.SigninInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.authService.Signin(r.Context(), &input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUserExists):
		writeError(w, http.StatusConflict, "user already exists")
	case errors.Is(err, auth.ErrInvalidUsername), errors.Is(err, auth.ErrInvalidFirstName), errors.Is(err, auth.ErrInvalidLastName), errors.Is(err, auth.ErrPasswordTooShort):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, auth.ErrInvalidPassword):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
