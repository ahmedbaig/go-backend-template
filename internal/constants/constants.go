package constants

import "time"

const Name = "go-backend-template"

const (
	EnvLocalhost  = "localhost"
	EnvTesting    = "testing"
	EnvStaging    = "staging"
	EnvProduction = "production"
)

const (
	TransportHTTP = "http"
	TransportGRPC = "grpc"
	HeaderAPIKey  = "X-Api-Key"
)

const (
	ProbeDBHandler = "db-handler"
	ProbeRedis     = "redis"
	ProbeUsers     = "users"
)

const (
	KeySession       = "session:"
	KeyEventsStream  = "events:stream:"
	KeyEventsChannel = "events:ch:"
	CookieName       = "sid"
	CookiePath       = "/"
)

const (
	HTTPClientTimeout   = 30 * time.Second
	SSERetry            = 3 * time.Second
	CapabilitiesJob     = "db-capabilities"
	CapabilitiesJitter  = 30 * time.Second
	CapabilitiesTimeout = 10 * time.Second
)
