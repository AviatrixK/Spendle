package goal

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Handler handles HTTP requests for savings goals and month-end sweeps.
type Handler struct {
	service *Service
}

// NewHandler constructs a goal HTTP Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateGoalRequest defines JSON payload for creating a goal.
type CreateGoalRequest struct {
	Title              string  `json:"title"`
	TargetPaise        int64   `json:"target_paise"`
	Deadline           *string `json:"deadline"`
	MonthlyAmountPaise *int64  `json:"monthly_amount_paise"`
	Priority           *int    `json:"priority"`
}

// UpdateGoalRequest defines JSON payload for updating a goal.
type UpdateGoalRequest struct {
	Title              string  `json:"title"`
	TargetPaise        int64   `json:"target_paise"`
	Deadline           *string `json:"deadline"`
	MonthlyAmountPaise *int64  `json:"monthly_amount_paise"`
	Priority           int     `json:"priority"`
	Status             string  `json:"status"`
}

// CloseMonthRequest defines JSON payload for manual month closing.
type CloseMonthRequest struct {
	Month string `json:"month"` // Format: YYYY-MM
}

// Create handles POST /api/v1/goals.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateGoalRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := CreateInput{
		Title:              req.Title,
		TargetPaise:        req.TargetPaise,
		Deadline:           req.Deadline,
		MonthlyAmountPaise: req.MonthlyAmountPaise,
		Priority:           req.Priority,
	}

	created, err := h.service.CreateGoal(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) ||
			errors.Is(err, ErrInvalidTarget) ||
			errors.Is(err, ErrConflictingTimeline) ||
			errors.Is(err, ErrInvalidDeadline) ||
			errors.Is(err, ErrInvalidMonthlyAmount) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to create savings goal")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusCreated, created)
}

// List handles GET /api/v1/goals.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	goals, err := h.service.ListGoals(r.Context(), userID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to list savings goals")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, goals)
}

// Update handles PUT /api/v1/goals/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	goalID := chi.URLParam(r, "id")
	if goalID == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "goal id is required")
		return
	}

	var req UpdateGoalRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := UpdateInput{
		Title:              req.Title,
		TargetPaise:        req.TargetPaise,
		Deadline:           req.Deadline,
		MonthlyAmountPaise: req.MonthlyAmountPaise,
		Priority:           req.Priority,
		Status:             req.Status,
	}

	updated, err := h.service.UpdateGoal(r.Context(), userID, goalID, input)
	if err != nil {
		if errors.Is(err, ErrGoalNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "goal not found")
			return
		}
		if errors.Is(err, ErrInvalidTitle) ||
			errors.Is(err, ErrInvalidTarget) ||
			errors.Is(err, ErrConflictingTimeline) ||
			errors.Is(err, ErrInvalidDeadline) ||
			errors.Is(err, ErrInvalidMonthlyAmount) ||
			errors.Is(err, ErrInvalidPriority) ||
			errors.Is(err, ErrInvalidStatus) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to update savings goal")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/goals/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	goalID := chi.URLParam(r, "id")
	if goalID == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "goal id is required")
		return
	}

	if err := h.service.DeleteGoal(r.Context(), userID, goalID); err != nil {
		if errors.Is(err, ErrGoalNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "goal not found")
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to delete savings goal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CloseMonth handles POST /api/v1/months/close.
func (h *Handler) CloseMonth(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CloseMonthRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	monthStr := strings.TrimSpace(req.Month)
	if monthStr == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "month is required (format: YYYY-MM)")
		return
	}

	parsedDate, err := time.Parse("2006-01", monthStr)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid month format: must be YYYY-MM")
		return
	}

	monthDate := time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Reject months in the future
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if monthDate.After(currentMonth) {
		_ = httpx.WriteError(w, http.StatusBadRequest, "cannot close a future month")
		return
	}

	result, err := h.service.CloseMonth(r.Context(), userID, monthDate)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to close month")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, result)
}

