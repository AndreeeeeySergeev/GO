package handler

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"bank-service/internal/middleware"
	"bank-service/internal/service"
)

type Handler struct {
	services *service.Services
	logger   *logrus.Logger
}

func NewHandler(services *service.Services, logger *logrus.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	public := router.PathPrefix("").Subrouter()
	h.registerPublicRoutes(public)

	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(h.services.User))
	h.registerProtectedRoutes(protected)
}

func (h *Handler) registerPublicRoutes(router *mux.Router) {
	router.HandleFunc("/register", h.Register).Methods("POST")
	router.HandleFunc("/login", h.Login).Methods("POST")
}

func (h *Handler) registerProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts", h.GetUserAccounts).Methods("GET")
	router.HandleFunc("/accounts/{id:[0-9]+}", h.GetAccount).Methods("GET")
	router.HandleFunc("/accounts/deposit", h.DepositToAccount).Methods("POST")
	router.HandleFunc("/accounts/withdraw", h.WithdrawFromAccount).Methods("POST")
	router.HandleFunc("/accounts/{id:[0-9]+}/predict", h.PredictBalance).Methods("GET")

	router.HandleFunc("/transfer", h.TransferFunds).Methods("POST")

	router.HandleFunc("/cards", h.CreateCard).Methods("POST")
	router.HandleFunc("/cards", h.GetUserCards).Methods("GET")
	router.HandleFunc("/cards/{id:[0-9]+}", h.GetCard).Methods("GET")
	router.HandleFunc("/cards/{id:[0-9]+}/status", h.UpdateCardStatus).Methods("PUT")
	router.HandleFunc("/cards/payment", h.ProcessCardPayment).Methods("POST")

	router.HandleFunc("/credits", h.ApplyForCredit).Methods("POST")
	router.HandleFunc("/credits", h.GetUserCredits).Methods("GET")
	router.HandleFunc("/credits/{id:[0-9]+}", h.GetCredit).Methods("GET")
	router.HandleFunc("/credits/{id:[0-9]+}/schedule", h.GetCreditSchedule).Methods("GET")

	router.HandleFunc("/transactions", h.GetUserTransactions).Methods("GET")
	router.HandleFunc("/accounts/{id:[0-9]+}/transactions", h.GetAccountTransactions).Methods("GET")

	router.HandleFunc("/analytics/transactions", h.GetTransactionAnalytics).Methods("GET")
	router.HandleFunc("/analytics/credits", h.GetCreditAnalytics).Methods("GET")
}
