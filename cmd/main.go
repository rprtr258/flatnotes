package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal/config"
	"github.com/rprtr258/flatnotes/internal/infra"
)

func run(ctx context.Context) error {
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	return infra.Run(ctx, cfg)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal().Err(err).Msg("app stopped")
	}
}
