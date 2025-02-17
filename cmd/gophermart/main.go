package main

import (
	"fmt"
	"log"

	"github.com/apetsko/gophermart/internal/config"
	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/storage"
	"github.com/apetsko/gophermart/internal/storage/postgres"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/server"
)

func main() {
	logger, err := logging.NewZapLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	cfg, err := config.Parse()
	if err != nil {
		logger.Fatal(err.Error())
	}

	st, err := storage.Init(cfg.DatabaseDSN, cfg.FileStoragePath, logger)
	if err != nil || st == nil {
		logger.Fatal(err.Error())
	}

	defer func(storage postgres.Storage) {
		err := storage.Close()
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to close storage: %s", err.Error()))
		}
	}(*st)

	handler := handlers.New(cfg.BaseURL, st, logger, cfg.Secret)

	//go storage.StartBatchDeleteProcessor(context.Background(), st, handler.ToDelete, logger)

	router := handlers.SetupRouter(handler)
	s := server.New(cfg.Host, router)

	logger.Info("running server on " + cfg.Host)
	if err := s.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}
