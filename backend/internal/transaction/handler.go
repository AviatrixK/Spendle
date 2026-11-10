package transaction

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Handler handles HTTP requests for financial transactions.
type Handler struct {
	service *Service
}

// NewHandler constructs a transaction HTTP Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateRequest defines the JSON payload for recording a new transaction.
type CreateRequest struct {
	Type        string  `json:"type"`
	AmountPaise int64   `json:"amount_paise"`
	CategoryID  *string `json:"category_id"`
	PaymentMode string  `json:"payment_mode"`
	Note        string  `json:"note"`
	OccurredOn  string  `json:"occurred_on"`
}

// Create handles POST /api/v1/transactions.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := CreateInput{
		Type:        req.Type,
		AmountPaise: req.AmountPaise,
		CategoryID:  req.CategoryID,
		PaymentMode: req.PaymentMode,
		Note:        req.Note,
		OccurredOn:  req.OccurredOn,
	}

	created, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, ErrExpenseCategoryRequired) ||
			errors.Is(err, ErrInvalidType) ||
			errors.Is(err, ErrInvalidAmount) ||
			errors.Is(err, ErrInvalidPaymentMode) ||
			errors.Is(err, ErrNoteTooLong) ||
			errors.Is(err, ErrInvalidOccurredOn) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to record transaction")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusCreated, created)
}

// List handles GET /api/v1/transactions?month=YYYY-MM&limit=50.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	monthStr := r.URL.Query().Get("month")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	transactions, err := h.service.List(r.Context(), userID, monthStr, limit)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, transactions)
}

// Delete handles DELETE /api/v1/transactions/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "transaction id is required")
		return
	}

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSummary handles GET /api/v1/transactions/summary?month=YYYY-MM.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	monthStr := r.URL.Query().Get("month")
	summary, err := h.service.GetSummary(r.Context(), userID, monthStr)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to get transaction summary")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, summary)
}

