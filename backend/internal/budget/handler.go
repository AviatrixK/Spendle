package budget

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Handler handles HTTP requests for categories and monthly budgets.
type Handler struct {
	service *Service
}

// NewHandler constructs a budget HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateCategoryRequest defines JSON payload for creating a custom category.
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

// SetBudgetRequest defines JSON payload for setting/upserting a monthly budget limit.
type SetBudgetRequest struct {
	CategoryID string `json:"category_id"`
	Month      string `json:"month"`
	LimitPaise int64  `json:"limit_paise"`
}

// ListCategories handles GET /api/v1/categories.
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	categories, err := h.service.ListCategories(r.Context(), userID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, categories)
}

// CreateCategory handles POST /api/v1/categories.
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateCategoryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.service.CreateCategory(r.Context(), userID, req.Name)
	if err != nil {
		if errors.Is(err, ErrCategoryExists) {
			_ = httpx.WriteError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidCategoryName) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusCreated, category)
}

// DeleteCategory handles DELETE /api/v1/categories/{id}.
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "category id is required")
		return
	}

	if err := h.service.DeleteCategory(r.Context(), userID, categoryID); err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "category not found")
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetBudget handles PUT /api/v1/budgets.
func (h *Handler) SetBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req SetBudgetRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	budget, err := h.service.SetBudget(r.Context(), userID, req.CategoryID, req.Month, req.LimitPaise)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, ErrInvalidLimit) || errors.Is(err, ErrInvalidMonth) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to set budget")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, budget)
}

// GetBudgets handles GET /api/v1/budgets?month=YYYY-MM.
// Go Concept - URL Query Parameters:
// In Go net/http, `r.URL.Query().Get("key")` extracts query parameters from the request URI.
// For example, in `/api/v1/budgets?month=2026-11`, calling `.Get("month")` yields `"2026-11"`.
func (h *Handler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	monthStr := r.URL.Query().Get("month")
	summary, err := h.service.GetMonthlyBudgets(r.Context(), userID, monthStr)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to retrieve monthly budgets")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, summary)
}

