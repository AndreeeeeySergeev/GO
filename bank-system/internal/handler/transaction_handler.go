package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"bank-service/internal/middleware"
	"bank-service/internal/service"
)

func (h *Handler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit, offset := getPaginationParams(r)

	transactions, err := h.services.Transaction.GetByUserID(userID, limit, offset)
	if err != nil {
		h.logger.Errorf("Failed to get user transactions: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	h.successResponse(w, http.StatusOK, transactions)
}

func (h *Handler) GetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	limit, offset := getPaginationParams(r)

	transactions, err := h.services.Transaction.GetByAccountID(accountID, userID, limit, offset)
	if err != nil {
		h.logger.Infof("Failed to get account transactions: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get transactions")
		}
		return
	}

	h.successResponse(w, http.StatusOK, transactions)
}

func getPaginationParams(r *http.Request) (limit, offset int) {
	limit = 10 // Значение по умолчанию
	offset = 0

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return
}
