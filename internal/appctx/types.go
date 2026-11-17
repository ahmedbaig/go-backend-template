package appctx

import (
	"log/slog"
	"net/http"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-monitoring"
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

	dbConn  *grpc.ClientConn
	closers []func() error
}
