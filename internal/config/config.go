package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	Accrual     string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Secret      string `env:"SECRET"`
}

func Parse() (c Config, err error) {
	flag.StringVar(&c.RunAddr, "a", ":8001", "network address with port")
	flag.StringVar(&c.DatabaseURI, "d", "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable", "database DSN")
	flag.StringVar(&c.Accrual, "r", "http://localhost:8080", "accrual system address")
	flag.StringVar(&c.Secret, "s", "42", "Secret")

	flag.Parse()

	if err = env.Parse(&c); err != nil {
		return c, fmt.Errorf("error while parse envs: %w", err)
	}
	return c, nil
}
