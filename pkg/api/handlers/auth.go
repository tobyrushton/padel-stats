package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tobyrushton/padel-stats/libs/auth"
)

type AuthService interface {
	Signup(ctx context.Context, input *auth.SignupInput) (*auth.AuthResult, error)
	Signin(ctx context.Context, input *auth.SigninInput) (*auth.AuthResult, error)
	SearchPlayers(ctx context.Context, query string) (*auth.SearchPlayersResult, error)
	ApproveUser(ctx context.Context, adminUserID, userID int64) (*auth.User, error)
}

type AuthHandler struct {
	authService      AuthService
	sessionValidator SessionValidator
}

func NewAuthHandler(authService AuthService, sessionValidator SessionValidator) *AuthHandler {
	return &AuthHandler{authService: authService, sessionValidator: sessionValidator}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.Signup)
		r.Post("/signin", h.Signin)
	})

	r.Route("/players", func(r chi.Router) {
		r.Get("/search", h.SearchPlayers)
	})

	r.Post("/admin/users/{userID}/approve", h.ApproveUser)
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
	fmt.Println(result, err)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SearchPlayers returns players matching a query.
// @Summary Search players
// @Description Search players by username, first name, or last name. Returns default player list when query is empty.
// @Tags players
// @Produce json
// @Param query query string false "Optional player search query"
// @Success 200 {object} auth.SearchPlayersResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/search [get]
func (h *AuthHandler) SearchPlayers(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))

	result, err := h.authService.SearchPlayers(r.Context(), query)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ApproveUser approves a pending user account.
// @Summary Approve user
// @Description Approve a pending user account. Requires an authenticated admin user.
// @Tags auth
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} auth.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/{userID}/approve [post]
func (h *AuthHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	tokenString := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if tokenString == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}

	session, err := h.sessionValidator.Validate(r.Context(), tokenString)
	if err != nil || session == nil || session.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	targetUserID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil || targetUserID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.authService.ApproveUser(r.Context(), session.UserID, targetUserID)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUserExists):
		writeError(w, http.StatusConflict, "user already exists")
	case errors.Is(err, auth.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, auth.ErrAdminAccessRequired):
		writeError(w, http.StatusForbidden, "admin access required")
	case errors.Is(err, auth.ErrUserPendingApproval):
		writeError(w, http.StatusForbidden, "user pending admin approval")
	case errors.Is(err, auth.ErrInvalidUsername), errors.Is(err, auth.ErrInvalidFirstName), errors.Is(err, auth.ErrInvalidLastName), errors.Is(err, auth.ErrPasswordTooShort):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, auth.ErrInvalidPassword):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
