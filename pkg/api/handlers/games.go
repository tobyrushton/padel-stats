package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	gamedomain "github.com/tobyrushton/padel-stats/libs/games"
)

type GamesService interface {
	CreateGame(ctx context.Context, creatorID int64, input *gamedomain.CreateGameInput) (*gamedomain.Game, error)
	ListGamesForPlayer(ctx context.Context, playerID int64) ([]*gamedomain.Game, error)
	GetGameByID(ctx context.Context, gameID int64) (*gamedomain.Game, error)
	DeleteGame(ctx context.Context, gameID int64) error
}

type SessionValidator interface {
	Validate(ctx context.Context, tokenString string) (*models.Session, error)
}

type GamesHandler struct {
	gamesService     GamesService
	sessionValidator SessionValidator
}

func NewGamesHandler(gamesService GamesService, sessionValidator SessionValidator) *GamesHandler {
	return &GamesHandler{gamesService: gamesService, sessionValidator: sessionValidator}
}

func (h *GamesHandler) RegisterRoutes(r chi.Router) {
	r.Route("/games", func(r chi.Router) {
		r.Post("/", h.CreateGame)
		r.Get("/{gameID}", h.GetGameByID)
		r.Delete("/{gameID}", h.DeleteGame)
	})

	r.Get("/players/{playerID}/games", h.ListGamesForPlayer)
}

// CreateGame creates a new game with four players.
// @Summary Create game
// @Description Create a game with two teams of two players and a final score.
// @Tags games
// @Accept json
// @Produce json
// @Param request body games.CreateGameInput true "Create game payload"
// @Success 201 {object} games.Game
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games [post]
func (h *GamesHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	tokenString := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if tokenString == "" {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	session, err := h.sessionValidator.Validate(r.Context(), tokenString)
	if err != nil || session == nil || session.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	var input gamedomain.CreateGameInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.gamesService.CreateGame(r.Context(), session.UserID, &input)
	if err != nil {
		handleGameError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetGameByID returns a game by ID including player details.
// @Summary Get game
// @Description Retrieve a game by its identifier.
// @Tags games
// @Produce json
// @Param gameID path int true "Game ID"
// @Success 200 {object} games.Game
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{gameID} [get]
func (h *GamesHandler) GetGameByID(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.ParseInt(chi.URLParam(r, "gameID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	result, err := h.gamesService.GetGameByID(r.Context(), gameID)
	if err != nil {
		fmt.Println(err)
		handleGameError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListGamesForPlayer returns all games for a player.
// @Summary List player games
// @Description Retrieve all games where the player participated.
// @Tags games
// @Produce json
// @Param playerID path int true "Player ID"
// @Success 200 {array} games.Game
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{playerID}/games [get]
func (h *GamesHandler) ListGamesForPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}

	result, err := h.gamesService.ListGamesForPlayer(r.Context(), playerID)
	if err != nil {
		handleGameError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteGame deletes a game by ID.
// @Summary Delete game
// @Description Delete a game by its identifier.
// @Tags games
// @Param gameID path int true "Game ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{gameID} [delete]
func (h *GamesHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.ParseInt(chi.URLParam(r, "gameID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	err = h.gamesService.DeleteGame(r.Context(), gameID)
	if err != nil {
		handleGameError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleGameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gamedomain.ErrGameNotFound):
		writeError(w, http.StatusNotFound, "game not found")
	case errors.Is(err, gamedomain.ErrNoSeasonForPlayedAt),
		errors.Is(err, gamedomain.ErrInvalidPlayerID),
		errors.Is(err, gamedomain.ErrDuplicatePlayers),
		errors.Is(err, gamedomain.ErrInvalidScore),
		errors.Is(err, gamedomain.ErrInvalidPlayedAt),
		errors.Is(err, gamedomain.ErrInvalidCreatorID),
		errors.Is(err, gamedomain.ErrInvalidGameID),
		errors.Is(err, gamedomain.ErrInvalidDeleteGame),
		errors.Is(err, gamedomain.ErrInvalidPlayerQuery):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, gamedomain.ErrSeasonOverlap):
		writeError(w, http.StatusInternalServerError, "internal server error")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func bearerTokenFromHeader(authorization string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(authorization, prefix) {
		return ""
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	if token == "" {
		return ""
	}

	return token
}
