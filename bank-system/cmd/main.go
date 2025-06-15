package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"bank-service/internal/config"
	"bank-service/internal/handler"
	"bank-service/internal/middleware"
	"bank-service/internal/repository"
	"bank-service/internal/scheduler"
	"bank-service/internal/service"
	"bank-service/pkg/logger"
)

func main() {
	log := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repos := repository.NewRepositories(db)

	encryptionService := service.NewEncryptionService(cfg)
	emailService := service.NewEmailService(cfg.SMTP)
	cbrService := service.NewCBRService()

	services := service.NewServices(service.Dependencies{
		Repos:             repos,
		EncryptionService: encryptionService,
		EmailService:      emailService,
		CBRService:        cbrService,
		Config:            cfg,
	})

	handlers := handler.NewHandler(services, log)

	router := mux.NewRouter()

	router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.RecoveryMiddleware(log))

	handlers.RegisterRoutes(router)

	creditScheduler := scheduler.NewCreditScheduler(services.Credit, log)
	go creditScheduler.Start(12 * time.Hour) // Проверка каждые 12 часов

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Infof("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	creditScheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")
}
