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

func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.CardCreation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	card, err := h.services.Card.Create(userID, input)
	if err != nil {
		h.logger.Infof("Failed to create card: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to create card")
		}
		return
	}

	h.logger.Infof("Card created successfully for user %d, account %d", userID, input.AccountID)
	h.successResponse(w, http.StatusCreated, card)
}

func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid card ID")
		return
	}

	card, err := h.services.Card.GetByID(cardID, userID)
	if err != nil {
		h.logger.Infof("Failed to get card: %v", err)

		switch err {
		case service.ErrCardNotFound:
			h.errorResponse(w, http.StatusNotFound, "Card not found")
		case service.ErrCardAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this card is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get card")
		}
		return
	}

	h.successResponse(w, http.StatusOK, card)
}

func (h *Handler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cards, err := h.services.Card.GetByUserID(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user cards: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get cards")
		return
	}

	h.successResponse(w, http.StatusOK, cards)
}

func (h *Handler) UpdateCardStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid card ID")
		return
	}

	var input struct {
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.services.Card.UpdateStatus(cardID, input.IsActive, userID); err != nil {
		h.logger.Infof("Failed to update card status: %v", err)

		switch err {
		case service.ErrCardNotFound:
			h.errorResponse(w, http.StatusNotFound, "Card not found")
		case service.ErrCardAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this card is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to update card status")
		}
		return
	}

	h.logger.Infof("Card status updated successfully: card %d, active: %v", cardID, input.IsActive)
	h.successResponse(w, http.StatusOK, map[string]string{"message": "Card status updated successfully"})
}

func (h *Handler) ProcessCardPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.CardPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.services.Card.ProcessPayment(input, userID); err != nil {
		h.logger.Infof("Failed to process payment: %v", err)

		switch err {
		case service.ErrCardNotFound:
			h.errorResponse(w, http.StatusNotFound, "Card not found")
		case service.ErrCardAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this card is denied")
		case service.ErrCardInactive:
			h.errorResponse(w, http.StatusBadRequest, "Card is inactive")
		case service.ErrInsufficientFunds:
			h.errorResponse(w, http.StatusBadRequest, "Insufficient funds")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to process payment")
		}
		return
	}

	h.logger.Infof("Payment processed successfully: %.2f using card %d", input.Amount, input.CardID)
	h.successResponse(w, http.StatusOK, map[string]string{"message": "Payment processed successfully"})
}
