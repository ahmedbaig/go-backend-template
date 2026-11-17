package config

import (
	cfg "github.com/ahmedbaig/go-backend-config"
	"github.com/ahmedbaig/go-backend-logging"
)

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
	Version         string       `json:"version" validate:"required"`
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

type Services struct {
	Users string `json:"users" validate:"omitempty,url"`
}

type Config struct {
	App    App
	Logger logging.Config
}
