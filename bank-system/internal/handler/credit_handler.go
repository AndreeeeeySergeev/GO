package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"bank-service/internal/middleware"
	"bank-service/internal/models"
	"bank-service/internal/service"
)

func (h *Handler) ApplyForCredit(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.CreditApplication
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	credit, err := h.services.Credit.Apply(userID, input)
	if err != nil {
		h.logger.Infof("Failed to apply for credit: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		case service.ErrInvalidCreditAmount:
			h.errorResponse(w, http.StatusBadRequest, "Credit amount must be positive")
		case service.ErrInvalidCreditTerm:
			h.errorResponse(w, http.StatusBadRequest, "Credit term must be between 3 and 60 months")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to apply for credit")
		}
		return
	}

	h.logger.Infof("Credit applied successfully for user %d, amount %.2f", userID, input.Amount)
	h.successResponse(w, http.StatusCreated, credit)
}

func (h *Handler) GetCredit(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid credit ID")
		return
	}

	credit, err := h.services.Credit.GetByID(creditID, userID)
	if err != nil {
		h.logger.Infof("Failed to get credit: %v", err)

		switch err {
		case service.ErrCreditNotFound:
			h.errorResponse(w, http.StatusNotFound, "Credit not found")
		case service.ErrCreditAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this credit is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get credit")
		}
		return
	}

	h.successResponse(w, http.StatusOK, credit)
}

func (h *Handler) GetUserCredits(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	credits, err := h.services.Credit.GetByUserID(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user credits: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get credits")
		return
	}

	h.successResponse(w, http.StatusOK, credits)
}

func (h *Handler) GetCreditSchedule(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid credit ID")
		return
	}

	schedule, err := h.services.Credit.GetSchedule(creditID, userID)
	if err != nil {
		h.logger.Infof("Failed to get credit schedule: %v", err)

		switch err {
		case service.ErrCreditNotFound:
			h.errorResponse(w, http.StatusNotFound, "Credit not found")
		case service.ErrCreditAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this credit is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get credit schedule")
		}
		return
	}

	h.successResponse(w, http.StatusOK, schedule)
}
