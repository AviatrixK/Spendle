package dashboard

import (
	"net/http"
	"strings"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Handler handles HTTP requests for the user dashboard.
type Handler struct {
	service *Service
}

// NewHandler constructs a dashboard HTTP Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetDashboard handles GET /api/v1/dashboard?month=YYYY-MM.
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	monthStr := strings.TrimSpace(r.URL.Query().Get("month"))
	if monthStr == "" {
		monthStr = time.Now().UTC().Format("2006-01")
	} else {
		if _, err := time.Parse("2006-01", monthStr); err != nil {
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid month format: must be YYYY-MM")
			return
		}
	}

	resp, err := h.service.GetDashboard(r.Context(), userID, monthStr)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, resp)
}

