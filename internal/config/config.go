package config

import (
	"net"
	"strconv"
	"time"

	cfg "github.com/ahmedbaig/go-backend-config"
	"github.com/ahmedbaig/go-backend-logging"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
)

func (a App) DevLoginEnabled() bool {
	return a.Session.DevLogin && (a.Environment == constants.EnvLocalhost || a.Environment == constants.EnvTesting)
}

func (a App) Addr() string { return net.JoinHostPort(a.Host, strconv.Itoa(a.Port)) }

func Defaults() App {
	return App{
		Version:         "0.0.0",
		Port:            8080,
		Host:            "0.0.0.0",
		Environment:     constants.EnvProduction,
		RequestTimeout:  cfg.Duration(30 * time.Second),
		ShutdownTimeout: cfg.Duration(15 * time.Second),
		DBHandler:       DBHandler{Transport: constants.TransportGRPC, GRPCAddr: "db-handler:9090", URL: "http://db-handler:8090", Timeout: cfg.Duration(30 * time.Second)},
		Redis:           Redis{Addr: "redis:6379", Prefix: constants.Name + ":"},
		Session:         Session{CookieName: constants.CookieName, Secure: true, TTL: cfg.Duration(7 * 24 * time.Hour)},
		Events:          Events{MaxLen: 1000, Heartbeat: cfg.Duration(15 * time.Second)},
		Jobs:            Jobs{CapabilitiesEvery: cfg.Duration(10 * time.Minute)},
	}
}

func Load() (Config, error) {
	app, err := cfg.Load(constants.Name, Defaults())
	if err != nil {
		return Config{}, err
	}
	log, err := cfg.Load(constants.Name+"-logger", logging.DefaultConfig())
	if err != nil {
		return Config{}, err
	}
	return Config{App: app, Logger: log}, nil
}
