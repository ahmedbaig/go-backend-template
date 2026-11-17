package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ahmedbaig/go-backend-logging"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/ahmedbaig/go-backend-template/internal/server"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := logging.New(cfg.Logger, constants.Name, logging.WithMeta(cfg.App.Environment, cfg.App.Version))
	if err != nil {
		return err
	}
	defer log.Close()
	slog.SetDefault(log.Logger)

	app, err := server.Build(cfg.App, log.Logger)
	if err != nil {
		return err
	}
	return app.Run(ctx, cfg.App.Addr())
}
