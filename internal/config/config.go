package config

import (
	"fmt"

	"github.com/apetsko/gophermart/internal/utils"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr     string `env:"GOPHERMART_RUN_ADDRESS" validate:"required"`
	DatabaseURI string `env:"GOPHERMART_DATABASE_URI" validate:"required"`
	Secret      string `env:"GOPHERMART_SECRET" validate:"required"`
	Accrual     string `env:"ACCRUAL_SYSTEM_ADDRESS" validate:"required,url"`
}

func New() (*Config, error) {
	var c Config

	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	if err := utils.ValidateStruct(c); err != nil {
		return nil, err
	}
	return &c, nil
}
