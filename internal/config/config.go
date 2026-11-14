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
	URL      string       `json:"url" validate:"required,url"`
	APIKey   string       `json:"apiKey"`
	Instance string       `json:"instance"`
	Timeout  cfg.Duration `json:"timeout"`
}

type App struct {
	Port            int          `json:"port" validate:"min=0,max=65535"`
	Host            string       `json:"host"`
	Environment     string       `json:"environment" validate:"oneof=localhost testing staging production"`
	RequestTimeout  cfg.Duration `json:"requestTimeout"`
	ShutdownTimeout cfg.Duration `json:"shutdownTimeout"`
	CORSOrigins     []string     `json:"corsOrigins"`
	DBHandler       DBHandler    `json:"dbHandler"`
	Services        Services     `json:"services"`
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
		DBHandler:       DBHandler{URL: "http://db-handler:8090", Timeout: cfg.Duration(30 * time.Second)},
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
