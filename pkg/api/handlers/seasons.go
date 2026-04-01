package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	seasonsdomain "github.com/tobyrushton/padel-stats/libs/seasons"
)

// SeasonsService defines season operations used by HTTP handlers.
type SeasonsService interface {
	GetSeasons(ctx context.Context) ([]*seasonsdomain.Season, error)
	GetActiveSeason(ctx context.Context) (*seasonsdomain.Season, error)
	CreateSeason(ctx context.Context, input *seasonsdomain.CreateSeasonInput) (*seasonsdomain.Season, error)
	EndSeason(ctx context.Context, seasonID int64) (*seasonsdomain.Season, error)
}

// SeasonsHandler handles HTTP requests for season resources.
type SeasonsHandler struct {
	seasonsService   SeasonsService
	sessionValidator SessionValidator
}

// NewSeasonsHandler creates a SeasonsHandler.
func NewSeasonsHandler(seasonsService SeasonsService, sessionValidator SessionValidator) *SeasonsHandler {
	return &SeasonsHandler{seasonsService: seasonsService, sessionValidator: sessionValidator}
}

// RegisterRoutes registers season routes.
func (h *SeasonsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/seasons", h.GetSeasons)
	r.Get("/seasons/active", h.GetActiveSeason)
	r.Post("/seasons", h.CreateSeason)
}

// GetSeasons returns all seasons.
// @Summary List seasons
// @Description Retrieve all seasons.
// @Tags seasons
// @Produce json
// @Success 200 {array} seasons.Season
// @Failure 500 {object} ErrorResponse
// @Router /seasons [get]
func (h *SeasonsHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	result, err := h.seasonsService.GetSeasons(r.Context())
	if err != nil {
		handleSeasonsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetActiveSeason returns the current active season.
// @Summary Get active season
// @Description Retrieve the currently active season.
// @Tags seasons
// @Produce json
// @Success 200 {object} seasons.Season
// @Failure 500 {object} ErrorResponse
// @Router /seasons/active [get]
func (h *SeasonsHandler) GetActiveSeason(w http.ResponseWriter, r *http.Request) {
	result, err := h.seasonsService.GetActiveSeason(r.Context())
	if err != nil {
		handleSeasonsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateSeason ends the current active season and creates a new one.
// @Summary Create season
// @Description Create a new season. Requires an authenticated admin user.
// @Tags seasons
// @Accept json
// @Produce json
// @Param request body seasons.CreateSeasonInput true "Create season payload"
// @Success 201 {object} seasons.Season
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons [post]
func (h *SeasonsHandler) CreateSeason(w http.ResponseWriter, r *http.Request) {
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

	if session.User.IsAdmin != true {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	var input seasonsdomain.CreateSeasonInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	currentSeason, err := h.seasonsService.GetActiveSeason(r.Context())
	if err != nil {
		handleSeasonsError(w, err)
		return
	}

	_, err = h.seasonsService.EndSeason(r.Context(), currentSeason.ID)
	if err != nil {
		handleSeasonsError(w, err)
		return
	}

	result, err := h.seasonsService.CreateSeason(r.Context(), &input)
	if err != nil {
		handleSeasonsError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func handleSeasonsError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusInternalServerError, "failed to retrieve seasons")
}
