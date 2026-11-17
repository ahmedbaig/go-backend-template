package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-logging"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/ahmedbaig/go-backend-types"
	"github.com/redis/go-redis/v9"
)

func NewStore(rdb redis.UniversalClient, o Options) *Store {
	if o.TTL <= 0 {
		o.TTL = 7 * 24 * time.Hour
	}
	if o.Cookie.Name == "" {
		o.Cookie.Name = constants.CookieName
	}
	return &Store{rdb: rdb, o: o}
}

func (s *Store) key(sid string) string { return s.o.Prefix + constants.KeySession + sid }

func newSID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Store) Create(ctx context.Context, u types.User) (string, error) {
	sid, err := newSID()
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(u)
	if err != nil {
		return "", err
	}
	return sid, s.rdb.Set(ctx, s.key(sid), b, s.o.TTL).Err()
}

func (s *Store) Get(ctx context.Context, sid string) (types.User, bool, error) {
	var u types.User
	b, err := s.rdb.GetEx(ctx, s.key(sid), s.o.TTL).Bytes()
	if errors.Is(err, redis.Nil) {
		return u, false, nil
	}
	if err != nil {
		return u, false, err
	}
	return u, true, json.Unmarshal(b, &u)
}

func (s *Store) Delete(ctx context.Context, sid string) error {
	return s.rdb.Del(ctx, s.key(sid)).Err()
}

func (s *Store) SetCookie(w http.ResponseWriter, sid string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.o.Cookie.Name,
		Value:    sid,
		Path:     constants.CookiePath,
		Domain:   s.o.Cookie.Domain,
		MaxAge:   int(s.o.TTL.Seconds()),
		HttpOnly: true,
		Secure:   s.o.Cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Store) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.o.Cookie.Name,
		Value:    "",
		Path:     constants.CookiePath,
		Domain:   s.o.Cookie.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.o.Cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Store) SID(r *http.Request) string {
	c, err := r.Cookie(s.o.Cookie.Name)
	if err != nil {
		return ""
	}
	return c.Value
}

func (s *Store) Load() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sid := s.SID(r)
			if sid == "" {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, true)
			u, ok, err := s.Get(ctx, sid)
			if err != nil {
				httpx.WriteError(w, r, httpx.E(httpx.CodeUnavailable).WithMessage("Sessions are unavailable").Wrap(err))
				return
			}
			if ok {
				ctx = types.WithUser(ctx, u)
				ctx = logging.WithAttrs(ctx, slog.Int64("userId", u.ID), slog.String("email", u.Email), slog.Any("roles", u.Roles))
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Require() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := types.UserFrom(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}
			code := httpx.CodeUnauthenticated
			if had, _ := r.Context().Value(ctxKey{}).(bool); had {
				code = httpx.CodeSessionExpired
			}
			httpx.WriteError(w, r, httpx.E(code))
		})
	}
}
