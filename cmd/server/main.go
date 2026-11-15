package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-logging"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/http/routes"
	"github.com/ahmedbaig/go-backend-template/internal/jobs"
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
	actx, err := appctx.New(c, log, version)
	if err != nil {
		return nil, err
	}
	if err := actx.RegisterProbes(); err != nil {
		_ = actx.Close(context.Background())
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
	app.OnShutdown(actx.Close)

	runner := jobs.NewRunner(log, jobs.Job{
		Name:     "db-capabilities",
		Every:    c.Jobs.CapabilitiesEvery.Std(),
		Jitter:   30 * time.Second,
		Timeout:  10 * time.Second,
		RunFirst: true,
		Run: func(ctx context.Context) error {
			caps, err := actx.DB.Capabilities(ctx)
			if err != nil {
				return err
			}
			log.InfoContext(ctx, "db-handler capabilities", slog.Int("databases", len(caps.Databases)), slog.Int("stores", len(caps.Stores)))
			return nil
		},
	})
	runner.Start(context.Background())
	app.OnShutdown(runner.Stop)
	return app, nil
}
