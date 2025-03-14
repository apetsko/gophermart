package main

import (
	"context"
	"log"

	"github.com/apetsko/gophermart/internal/clients/accrual"
	"github.com/apetsko/gophermart/internal/config"
	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"go.uber.org/zap/zapcore"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/server"
)

func main() {
	logger, err := logging.NewLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	logger.Infof("Starting server with LogLevel: %s", zapcore.DebugLevel)

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(err.Error())
	}

	st, err := postgres.New(cfg.DatabaseURI, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	defer func() {
		if err := st.Close(); err != nil {
			logger.Error("Failed to close storage:", err)
		}
	}()

	handler := handlers.New(st, cfg.Secret, logger)
	router := handlers.SetupRouter(handler)
	s := server.New(cfg.RunAddr, router)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCount := 5
	accrual.ProcessAccrual(ctx, st.DB, cfg.Accrual, workerCount, logger)

	logger.Info("Running server on " + cfg.RunAddr)

	if err := s.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}
