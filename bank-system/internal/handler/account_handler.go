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

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Type models.AccountType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	account, err := h.services.Account.Create(userID, input.Type)
	if err != nil {
		h.logger.Errorf("Failed to create account: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	h.logger.Infof("Account created successfully for user %d", userID)
	h.successResponse(w, http.StatusCreated, account)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
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

	account, err := h.services.Account.GetByID(accountID, userID)
	if err != nil {
		h.logger.Infof("Failed to get account: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get account")
		}
		return
	}

	h.successResponse(w, http.StatusOK, account)
}

func (h *Handler) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accounts, err := h.services.Account.GetByUserID(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user accounts: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get accounts")
		return
	}

	h.successResponse(w, http.StatusOK, accounts)
}

func (h *Handler) DepositToAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.services.Account.Deposit(input, userID); err != nil {
		h.logger.Infof("Failed to deposit: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		case service.ErrInvalidAmount:
			h.errorResponse(w, http.StatusBadRequest, "Amount must be positive")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to deposit")
		}
		return
	}

	h.logger.Infof("Deposit successful: %.2f to account %d", input.Amount, input.AccountID)
	h.successResponse(w, http.StatusOK, map[string]string{"message": "Deposit successful"})
}

func (h *Handler) WithdrawFromAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.services.Account.Withdraw(input, userID); err != nil {
		h.logger.Infof("Failed to withdraw: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		case service.ErrInvalidAmount:
			h.errorResponse(w, http.StatusBadRequest, "Amount must be positive")
		case service.ErrInsufficientFunds:
			h.errorResponse(w, http.StatusBadRequest, "Insufficient funds")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to withdraw")
		}
		return
	}

	h.logger.Infof("Withdrawal successful: %.2f from account %d", input.Amount, input.AccountID)
	h.successResponse(w, http.StatusOK, map[string]string{"message": "Withdrawal successful"})
}

func (h *Handler) TransferFunds(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.services.Account.Transfer(input, userID); err != nil {
		h.logger.Infof("Failed to transfer: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to source account is denied")
		case service.ErrInvalidAmount:
			h.errorResponse(w, http.StatusBadRequest, "Amount must be positive")
		case service.ErrInsufficientFunds:
			h.errorResponse(w, http.StatusBadRequest, "Insufficient funds")
		case service.ErrSameAccount:
			h.errorResponse(w, http.StatusBadRequest, "Cannot transfer to the same account")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to transfer")
		}
		return
	}

	h.logger.Infof("Transfer successful: %.2f from account %d to account %d",
		input.Amount, input.FromAccountID, input.ToAccountID)
	h.successResponse(w, http.StatusOK, map[string]string{"message": "Transfer successful"})
}

func (h *Handler) PredictBalance(w http.ResponseWriter, r *http.Request) {
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

	days := 30 // Значение по умолчанию
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	predictions, err := h.services.Account.PredictBalance(accountID, userID, days)
	if err != nil {
		h.logger.Infof("Failed to predict balance: %v", err)

		switch err {
		case service.ErrAccountNotFound:
			h.errorResponse(w, http.StatusNotFound, "Account not found")
		case service.ErrAccountAccessDenied:
			h.errorResponse(w, http.StatusForbidden, "Access to this account is denied")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to predict balance")
		}
		return
	}

	h.successResponse(w, http.StatusOK, predictions)
}
