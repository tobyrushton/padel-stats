package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	leaderboarddomain "github.com/tobyrushton/padel-stats/libs/leaderboard"
)

type LeaderboardService interface {
	GetSeasonLeaderboard(ctx context.Context, seasonID int64) ([]*leaderboarddomain.Entry, error)
	GetAllTimeLeaderboard(ctx context.Context) ([]*leaderboarddomain.Entry, error)
}

type LeaderboardHandler struct {
	leaderboardService LeaderboardService
}

func NewLeaderboardHandler(leaderboardService LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

func (h *LeaderboardHandler) RegisterRoutes(r chi.Router) {
	r.Get("/seasons/{seasonID}/leaderboard", h.GetSeasonLeaderboard)
	r.Get("/leaderboard", h.GetAllTimeLeaderboard)
}

// GetSeasonLeaderboard returns leaderboard rows for a season.
// @Summary Get season leaderboard
// @Description Retrieve leaderboard standings derived from game score difference for one season.
// @Tags leaderboard
// @Produce json
// @Param seasonID path int true "Season ID"
// @Success 200 {array} leaderboard.Entry
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{seasonID}/leaderboard [get]
func (h *LeaderboardHandler) GetSeasonLeaderboard(w http.ResponseWriter, r *http.Request) {
	seasonID, err := strconv.ParseInt(chi.URLParam(r, "seasonID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season id")
		return
	}

	result, err := h.leaderboardService.GetSeasonLeaderboard(r.Context(), seasonID)
	if err != nil {
		handleLeaderboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetAllTimeLeaderboard returns all-time leaderboard rows.
// @Summary Get all-time leaderboard
// @Description Retrieve leaderboard standings derived from game score difference across all seasons.
// @Tags leaderboard
// @Produce json
// @Success 200 {array} leaderboard.Entry
// @Failure 500 {object} ErrorResponse
// @Router /leaderboard [get]
func (h *LeaderboardHandler) GetAllTimeLeaderboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.leaderboardService.GetAllTimeLeaderboard(r.Context())
	if err != nil {
		handleLeaderboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func handleLeaderboardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, leaderboarddomain.ErrInvalidSeasonID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
