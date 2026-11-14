package appctx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-monitoring"
	"github.com/ahmedbaig/go-backend-template/internal/config"
)

type Context struct {
	Config  config.App
	Log     *slog.Logger
	Prober  *monitoring.Prober
	DB      *db.Client
	HTTP    *http.Client
	Version string
}

func New(c config.App, log *slog.Logger, version string) *Context {
	httpClient := &http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: 30 * time.Second}
	dbHTTP := &http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: c.DBHandler.Timeout.Std()}
	opts := []db.ClientOption{db.WithHTTPClient(dbHTTP)}
	if c.DBHandler.APIKey != "" {
		opts = append(opts, db.WithHeader("X-Api-Key", c.DBHandler.APIKey))
	}
	if c.DBHandler.Instance != "" {
		opts = append(opts, db.WithInstance(c.DBHandler.Instance))
	}
	return &Context{
		Config:  c,
		Log:     log,
		Prober:  monitoring.NewProber(config.Name, monitoring.WithVersion(version)),
		DB:      db.NewClient(c.DBHandler.URL, opts...),
		HTTP:    httpClient,
		Version: version,
	}
}

func (c *Context) RegisterProbes() error {
	if err := c.Prober.AddServiceProbe("db-handler", c.Config.DBHandler.URL); err != nil {
		return err
	}
	if c.Config.Services.Users != "" {
		return c.Prober.AddServiceProbe("users", c.Config.Services.Users, monitoring.Optional())
	}
	return nil
}
