package config

import (
	"net"
	"strconv"
	"time"

	cfg "github.com/ahmedbaig/go-backend-config"
	"github.com/ahmedbaig/go-backend-logging"
)

const Name = "go-backend-template"

type DBHandler struct {
	Transport string       `json:"transport" validate:"oneof=grpc http"`
	URL       string       `json:"url" validate:"required_if=Transport http,omitempty,url"`
	GRPCAddr  string       `json:"grpcAddr" validate:"required_if=Transport grpc"`
	APIKey    string       `json:"apiKey"`
	Instance  string       `json:"instance"`
	Timeout   cfg.Duration `json:"timeout"`
}

type Redis struct {
	Addr     string `json:"addr" validate:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
	DB       int    `json:"db" validate:"min=0"`
	TLS      bool   `json:"tls"`
	Prefix   string `json:"prefix"`
}

type Session struct {
	CookieName   string       `json:"cookieName"`
	CookieDomain string       `json:"cookieDomain"`
	Secure       bool         `json:"secure"`
	TTL          cfg.Duration `json:"ttl"`
	DevLogin     bool         `json:"devLogin"`
}

type Jobs struct {
	CapabilitiesEvery cfg.Duration `json:"capabilitiesEvery"`
}

type Events struct {
	MaxLen    int64        `json:"maxLen" validate:"min=0"`
	Heartbeat cfg.Duration `json:"heartbeat"`
}

type App struct {
	Port            int          `json:"port" validate:"min=0,max=65535"`
	Host            string       `json:"host"`
	Environment     string       `json:"environment" validate:"oneof=localhost testing staging production"`
	RequestTimeout  cfg.Duration `json:"requestTimeout"`
	ShutdownTimeout cfg.Duration `json:"shutdownTimeout"`
	CORSOrigins     []string     `json:"corsOrigins"`
	DBHandler       DBHandler    `json:"dbHandler"`
	Redis           Redis        `json:"redis"`
	Session         Session      `json:"session"`
	Events          Events       `json:"events"`
	Jobs            Jobs         `json:"jobs"`
	Services        Services     `json:"services"`
}

func (a App) DevLoginEnabled() bool {
	return a.Session.DevLogin && (a.Environment == "localhost" || a.Environment == "testing")
}

type Services struct {
	Users string `json:"users" validate:"omitempty,url"`
}

func (a App) Addr() string { return net.JoinHostPort(a.Host, strconv.Itoa(a.Port)) }

func Defaults() App {
	return App{
		Port:            8080,
		Host:            "0.0.0.0",
		Environment:     "production",
		RequestTimeout:  cfg.Duration(30 * time.Second),
		ShutdownTimeout: cfg.Duration(15 * time.Second),
		DBHandler:       DBHandler{Transport: "grpc", GRPCAddr: "db-handler:9090", URL: "http://db-handler:8090", Timeout: cfg.Duration(30 * time.Second)},
		Redis:           Redis{Addr: "redis:6379", Prefix: Name + ":"},
		Session:         Session{CookieName: "sid", Secure: true, TTL: cfg.Duration(7 * 24 * time.Hour)},
		Events:          Events{MaxLen: 1000, Heartbeat: cfg.Duration(15 * time.Second)},
		Jobs:            Jobs{CapabilitiesEvery: cfg.Duration(10 * time.Minute)},
	}
}

type Config struct {
	App    App
	Logger logging.Config
}

func Load() (Config, error) {
	app, err := cfg.Load(Name, Defaults())
	if err != nil {
		return Config{}, err
	}
	log, err := cfg.Load(Name+"-logger", logging.DefaultConfig())
	if err != nil {
		return Config{}, err
	}
	return Config{App: app, Logger: log}, nil
}
