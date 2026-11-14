package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-logging"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/http/routes"
)

var version = "dev"

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
	log, err := logging.New(cfg.Logger, config.Name)
	if err != nil {
		return err
	}
	defer log.Close()
	slog.SetDefault(log.Logger)

	app, err := build(cfg.App, log.Logger)
	if err != nil {
		return err
	}
	return app.Run(ctx, cfg.App.Addr())
}

func build(c config.App, log *slog.Logger) (*httpx.App, error) {
	actx := appctx.New(c, log, version)
	if err := actx.RegisterProbes(); err != nil {
		return nil, err
	}
	opts := httpx.Options{
		Name:            config.Name,
		Environment:     c.Environment,
		Logger:          log,
		Prober:          actx.Prober,
		RequestTimeout:  c.RequestTimeout.Std(),
		ShutdownTimeout: c.ShutdownTimeout.Std(),
		RateLimit:       &httpx.RateLimitOptions{},
	}
	if len(c.CORSOrigins) > 0 {
		opts.CORS = &httpx.CORSOptions{AllowedOrigins: c.CORSOrigins, AllowCredentials: true}
	}
	app := httpx.New(opts)
	routes.Register(app, actx)
	return app, nil
}
