package controllers

import (
	"net/http"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-http/sse"
	"github.com/ahmedbaig/go-backend-template/internal/events"
)

type Events struct {
	broker    *events.Broker
	topics    events.Topics
	heartbeat time.Duration
	draining  <-chan struct{}
}

func NewEvents(b *events.Broker, t events.Topics, heartbeat time.Duration, draining <-chan struct{}) *Events {
	return &Events{broker: b, topics: t, heartbeat: heartbeat, draining: draining}
}

func (c *Events) Stream(w http.ResponseWriter, r *http.Request) error {
	u, err := httpx.User(r.Context())
	if err != nil {
		return err
	}
	channels, err := c.topics.Resolve(u, r.URL.Query().Get("topics"))
	if err != nil {
		return err
	}
	ch, err := c.broker.Subscribe(r.Context(), channels, sse.LastEventID(r))
	if err != nil {
		return httpx.E(httpx.CodeUnavailable).WithMessage("Live updates are unavailable").Wrap(err)
	}
	s, err := sse.Start(w)
	if err != nil {
		return err
	}
	return sse.Serve(r.Context(), s, ch, sse.Options{Heartbeat: c.heartbeat, Retry: 3 * time.Second, Done: c.draining})
}
