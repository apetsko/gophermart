package main

import (
	"fmt"
	"log"

	"github.com/apetsko/gophermart/internal/accrual"
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

	st, err := storage.Init(cfg.DatabaseURI, logger)
	if err != nil || st == nil {
		logger.Fatal(err.Error())
	}

	defer func(storage postgres.Storage) {
		err := storage.Close()
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to close storage: %s", err.Error()))
		}
	}(*st)

	handler := handlers.New(st, logger)

	//go storage.StartBatchDeleteProcessor(context.Background(), st, handler.ToDelete, logger)

	router := handlers.SetupRouter(handler)
	s := server.New(cfg.RunAddr, router)

	if err := accrual.InitAccrualForTests(logger); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("running server on " + cfg.RunAddr)
	if err := s.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}
