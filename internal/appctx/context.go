package appctx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-monitoring"
	dbv1 "github.com/ahmedbaig/go-backend-proto/gen/go/ahmedbaig/dbhandler/v1"
	"github.com/ahmedbaig/go-backend-proto/grpcx"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/ahmedbaig/go-backend-template/internal/events"
)

func New(c config.App, log *slog.Logger) (*Context, error) {
	ctx := &Context{
		Config: c,
		Log:    log,
		Prober: monitoring.NewProber(constants.Name, monitoring.WithVersion(c.Version)),
		HTTP:   &http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: constants.HTTPClientTimeout},
		Topics: events.NewTopics(events.Topic{Name: api.TopicPosts, Scope: api.ScopePostsRead, PerUser: true}),
	}
	var err error
	if ctx.DB, ctx.dbConn, err = connectDB(c.DBHandler); err != nil {
		return nil, err
	}
	if ctx.dbConn != nil {
		ctx.closers = append(ctx.closers, ctx.dbConn.Close)
	}
	ctx.Redis = connectRedis(c.Redis)
	ctx.closers = append(ctx.closers, ctx.Redis.Close)
	ctx.Sessions = newSessions(ctx.Redis, c)
	ctx.Events = events.NewBroker(ctx.Redis, events.Options{Prefix: c.Redis.Prefix, MaxLen: c.Events.MaxLen})
	return ctx, nil
}

func (c *Context) RegisterProbes() error {
	if c.dbConn != nil {
		if err := c.Prober.AddProbe(constants.ProbeDBHandler, grpcx.HealthCheck(c.dbConn, dbv1.DBHandlerService_ServiceDesc.ServiceName)); err != nil {
			return err
		}
	} else if err := c.Prober.AddServiceProbe(constants.ProbeDBHandler, c.Config.DBHandler.URL); err != nil {
		return err
	}
	if err := c.Prober.AddProbe(constants.ProbeRedis, func(ctx context.Context) error { return c.Redis.Ping(ctx).Err() }); err != nil {
		return err
	}
	if c.Config.Services.Users != "" {
		return c.Prober.AddServiceProbe(constants.ProbeUsers, c.Config.Services.Users, monitoring.Optional())
	}
	return nil
}

func (c *Context) Close(context.Context) error {
	var err error
	for i := len(c.closers) - 1; i >= 0; i-- {
		err = errors.Join(err, c.closers[i]())
	}
	c.closers = nil
	return err
}
