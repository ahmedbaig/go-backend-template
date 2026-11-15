package appctx

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-db/dbgrpc"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-monitoring"
	dbv1 "github.com/ahmedbaig/go-backend-proto/gen/go/ahmedbaig/dbhandler/v1"
	"github.com/ahmedbaig/go-backend-proto/grpcx"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/events"
	"github.com/ahmedbaig/go-backend-template/internal/session"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type Context struct {
	Config   config.App
	Log      *slog.Logger
	Prober   *monitoring.Prober
	DB       *db.Client
	HTTP     *http.Client
	Redis    redis.UniversalClient
	Sessions *session.Store
	Events   *events.Broker
	Topics   events.Topics
	Version  string

	dbConn  *grpc.ClientConn
	closers []func() error
}

func New(c config.App, log *slog.Logger, version string) (*Context, error) {
	ctx := &Context{
		Config:  c,
		Log:     log,
		Prober:  monitoring.NewProber(config.Name, monitoring.WithVersion(version)),
		HTTP:    &http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: 30 * time.Second},
		Version: version,
	}
	if err := ctx.openDB(); err != nil {
		return nil, err
	}
	ctx.openRedis()
	ctx.Topics = events.NewTopics(
		events.Topic{Name: api.TopicPosts, Scope: api.ScopePostsRead, PerUser: true},
	)
	return ctx, nil
}

func (c *Context) openDB() error {
	h := c.Config.DBHandler
	var opts []db.ClientOption
	if h.Instance != "" {
		opts = append(opts, db.WithInstance(h.Instance))
	}
	if h.Transport == "http" {
		opts = append(opts, db.WithHTTPClient(&http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: h.Timeout.Std()}))
		if h.APIKey != "" {
			opts = append(opts, db.WithHeader("X-Api-Key", h.APIKey))
		}
		c.DB = db.NewClient(h.URL, opts...)
		return nil
	}
	client, conn, err := dbgrpc.Dial(h.GRPCAddr, grpcx.ClientOptions{
		ServiceToken: h.APIKey,
		Extra:        []grpc.DialOption{grpc.WithUnaryInterceptor(timeout(h.Timeout.Std()))},
	}, opts...)
	if err != nil {
		return err
	}
	c.DB, c.dbConn = client, conn
	c.closers = append(c.closers, conn.Close)
	return nil
}

func timeout(d time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok && d > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Context) openRedis() {
	r := c.Config.Redis
	o := &redis.Options{Addr: r.Addr, Username: r.Username, Password: r.Password, DB: r.DB}
	if r.TLS {
		o.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	rdb := redis.NewClient(o)
	c.Redis = rdb
	c.closers = append(c.closers, rdb.Close)
	s := c.Config.Session
	c.Sessions = session.NewStore(rdb, session.Options{
		Prefix: r.Prefix,
		TTL:    s.TTL.Std(),
		Cookie: session.CookieOptions{Name: s.CookieName, Domain: s.CookieDomain, Secure: s.Secure},
	})
	c.Events = events.NewBroker(rdb, events.Options{Prefix: r.Prefix, MaxLen: c.Config.Events.MaxLen})
}

func (c *Context) RegisterProbes() error {
	if c.dbConn != nil {
		if err := c.Prober.AddProbe("db-handler", grpcx.HealthCheck(c.dbConn, dbv1.DBHandlerService_ServiceDesc.ServiceName)); err != nil {
			return err
		}
	} else if err := c.Prober.AddServiceProbe("db-handler", c.Config.DBHandler.URL); err != nil {
		return err
	}
	if err := c.Prober.AddProbe("redis", func(ctx context.Context) error { return c.Redis.Ping(ctx).Err() }); err != nil {
		return err
	}
	if c.Config.Services.Users != "" {
		return c.Prober.AddServiceProbe("users", c.Config.Services.Users, monitoring.Optional())
	}
	return nil
}

func (c *Context) Close(context.Context) error {
	var err error
	for i := len(c.closers) - 1; i >= 0; i-- {
		err = errors.Join(err, c.closers[i]())
	}
	return err
}
