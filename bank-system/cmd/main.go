package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"bank-system/internal/config"
	"bank-system/internal/handler"
	"bank-system/internal/middleware"
	"bank-system/internal/repository"
	"bank-system/internal/scheduler"
	"bank-system/internal/service"
	"bank-system/pkg/logger"
)

func main() {
	log := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
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

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Infof("Запуск сервера %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Выключение сервера...")

	creditScheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Сервер выключен принудительно: %v", err)
	}

	log.Info("Сервер выключен без ошибок")
}
