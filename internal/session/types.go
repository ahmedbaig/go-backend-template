package session

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Prefix string
	TTL    time.Duration
	Cookie CookieOptions
}

type CookieOptions struct {
	Name   string
	Domain string
	Secure bool
}

type Store struct {
	rdb redis.UniversalClient
	o   Options
}

type ctxKey struct{}
