package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

var JWTSecret = []byte ("my-super-secret-key-for-testing")

type Config struct {
	ServerAdress         string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

const (
	defaultServerAddress        = "localhost:8080"
	defaultDatabaseDSN          = "postgres://postgres:admin@localhost:5432/gophermart?sslmode=disable"
	defaultAccrualSystemAddress = "" //TODO
)

func Init() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.ServerAdress, "a", cfg.ServerAdress, "Server address")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.Parse()

	if cfg.ServerAdress == "" {
		cfg.ServerAdress = defaultServerAddress
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = defaultDatabaseDSN
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = defaultAccrualSystemAddress
	}

	return cfg, nil
}
