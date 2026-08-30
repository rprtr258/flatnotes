package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/flatnotes/internal/config"
	"github.com/rprtr258/flatnotes/internal/healthcheck"
)

func run() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	app, err := internal.New(cfg.DataPath)
	if err != nil {
		return fmt.Errorf("init app: %w", err)
	}

	healthcheck.Run(app)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("read config")
	}
}
