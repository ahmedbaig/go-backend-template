package server

import (
	"context"
	"log/slog"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/ahmedbaig/go-backend-template/internal/http/routes"
	"github.com/ahmedbaig/go-backend-template/internal/jobs"
)

func Build(c config.App, log *slog.Logger) (*httpx.App, error) {
	actx, err := appctx.New(c, log)
	if err != nil {
		return nil, err
	}
	if err := actx.RegisterProbes(); err != nil {
		_ = actx.Close(context.Background())
		return nil, err
	}
	opts := httpx.Options{
		Name:            constants.Name,
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

	runner := jobs.NewRunner(log, jobs.Capabilities(actx.DB, log, c.Jobs.CapabilitiesEvery.Std()))
	runner.Start(context.Background())
	app.OnShutdown(runner.Stop)
	return app, nil
}
