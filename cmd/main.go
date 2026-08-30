package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/rprtr258/flatnotes/internal/config"
	"github.com/rprtr258/flatnotes/internal/infra"
)

func run(ctx context.Context) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	return infra.Run(ctx, cfg)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.SetFlags(log.Lshortfile | log.Flags())
	if err := run(ctx); err != nil {
		log.Fatalf("app stopped: %s", err.Error())
	}
}
