package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/flatnotes/internal/healthcheck"
)

func run(args []string) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if len(args) != 1 {
		return fmt.Errorf("usage: check <dir>")
	}
	dir := args[0]

	app, err := internal.New(dir)
	if err != nil {
		return fmt.Errorf("init app: %w", err)
	}

	healthcheck.Run(app)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("read config")
	}
}
