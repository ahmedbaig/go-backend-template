package controllers

import (
	"hash/fnv"
	"net/http"
	"strings"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/session"
	"github.com/ahmedbaig/go-backend-types"
)

func NewSession(store *session.Store) *Session { return &Session{store: store} }

func (c *Session) Get(w http.ResponseWriter, r *http.Request) error {
	u, err := httpx.User(r.Context())
	if err != nil {
		return err
	}
	httpx.OK(w, api.SessionResponse{User: u})
	return nil
}

func (c *Session) DevLogin(w http.ResponseWriter, r *http.Request) error {
	in, err := httpx.Bind[api.DevLoginRequest](r)
	if err != nil {
		return err
	}
	u := types.User{ID: in.ID, Email: strings.ToLower(in.Email), Name: in.Name, Roles: in.Roles, Features: in.Features}
	if u.ID == 0 {
		u.ID = stableID(u.Email)
	}
	sid, err := c.store.Create(r.Context(), u)
	if err != nil {
		return httpx.E(httpx.CodeUnavailable).WithMessage("Sessions are unavailable").Wrap(err)
	}
	c.store.SetCookie(w, sid)
	httpx.LoggerFrom(r.Context()).InfoContext(r.Context(), "dev login", "userId", u.ID)
	httpx.OK(w, api.SessionResponse{User: u})
	return nil
}

func (c *Session) Logout(w http.ResponseWriter, r *http.Request) error {
	if sid := c.store.SID(r); sid != "" {
		if err := c.store.Delete(r.Context(), sid); err != nil {
			return httpx.E(httpx.CodeUnavailable).Wrap(err)
		}
	}
	c.store.ClearCookie(w)
	httpx.NoContent(w)
	return nil
}

func stableID(email string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(email))
	return int64(h.Sum64()>>12) + 1
}
