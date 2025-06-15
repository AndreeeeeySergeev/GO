package handler

import (
	"encoding/json"
	"net/http"

	"bank-service/internal/models"
	"bank-service/internal/service"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.UserRegistration

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Errorf("Failed to decode request body: %v", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := input.Validate(); err != nil {
		h.logger.Infof("Validation failed: %v", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.User.Register(input)
	if err != nil {
		h.logger.Errorf("Failed to register user: %v", err)

		switch err {
		case service.ErrUserExists:
			h.errorResponse(w, http.StatusConflict, "User with this email or username already exists")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	h.logger.Infof("User registered successfully: %s", user.Username)
	h.successResponse(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.UserLogin

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Errorf("Failed to decode request body: %v", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := h.services.User.Login(input)
	if err != nil {
		h.logger.Infof("Login failed: %v", err)

		switch err {
		case service.ErrInvalidCredentials:
			h.errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to authenticate")
		}
		return
	}

	h.logger.Infof("User logged in successfully: %s", input.Email)
	h.successResponse(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) successResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
