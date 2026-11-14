package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-logging"
)

const HeaderUser = "X-User"

type User struct {
	ID    int64    `json:"id"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

func (u User) Has(role string) bool { return slices.Contains(u.Roles, role) }

type userKey struct{}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

func RequireUser() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(HeaderUser)
			if raw == "" {
				httpx.WriteError(w, r, httpx.E(httpx.CodeUnauthenticated))
				return
			}
			var u User
			if err := json.Unmarshal([]byte(raw), &u); err != nil || u.ID == 0 {
				httpx.WriteError(w, r, httpx.E(httpx.CodeUnauthenticated).WithMessage("The session header is invalid"))
				return
			}
			ctx := WithUser(r.Context(), u)
			ctx = logging.WithAttrs(ctx, slog.Int64("userId", u.ID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFrom(r.Context())
			if !ok {
				httpx.WriteError(w, r, httpx.E(httpx.CodeUnauthenticated))
				return
			}
			for _, role := range roles {
				if u.Has(role) {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, r, httpx.E(httpx.CodeForbidden))
		})
	}
}
