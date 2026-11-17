package appctx

import (
	"crypto/tls"

	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/session"
	"github.com/redis/go-redis/v9"
)

func connectRedis(r config.Redis) redis.UniversalClient {
	o := &redis.Options{Addr: r.Addr, Username: r.Username, Password: r.Password, DB: r.DB}
	if r.TLS {
		o.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return redis.NewClient(o)
}

func newSessions(rdb redis.UniversalClient, c config.App) *session.Store {
	s := c.Session
	return session.NewStore(rdb, session.Options{
		Prefix: c.Redis.Prefix,
		TTL:    s.TTL.Std(),
		Cookie: session.CookieOptions{Name: s.CookieName, Domain: s.CookieDomain, Secure: s.Secure},
	})
}
