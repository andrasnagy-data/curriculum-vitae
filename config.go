package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	Env       string `env:"ENV"`
	Port      int    `env:"PORT"`
	SentryDSN string `env:"SENTRY_DSN"`
}

func NewConfig() *config {
	var cfg config
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal("error while reading '.env'")
	}

	return &cfg
}

func (c *config) Addr() string {
	return fmt.Sprintf("localhost:%d", c.Port)
}

func (c *config) IsProd() bool {
	return c.Env == "PROD"
}
