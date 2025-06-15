package handler

import (
	"net/http"

	"bank-service/internal/middleware"
)

func (h *Handler) GetTransactionAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // Значение по умолчанию
	}

	if period != "week" && period != "month" && period != "year" {
		h.errorResponse(w, http.StatusBadRequest, "Invalid period. Use 'week', 'month', or 'year'")
		return
	}

	analytics, err := h.services.Analytics.GetTransactionAnalytics(userID, period)
	if err != nil {
		h.logger.Errorf("Failed to get transaction analytics: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}

	h.successResponse(w, http.StatusOK, analytics)
}

func (h *Handler) GetCreditAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	analytics, err := h.services.Analytics.GetCreditAnalytics(userID)
	if err != nil {
		h.logger.Errorf("Failed to get credit analytics: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}

	h.successResponse(w, http.StatusOK, analytics)
}
