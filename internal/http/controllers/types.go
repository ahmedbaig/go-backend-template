package controllers

import (
	"time"

	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/events"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
	"github.com/ahmedbaig/go-backend-template/internal/session"
)

type Events struct {
	broker    *events.Broker
	topics    events.Topics
	heartbeat time.Duration
	draining  <-chan struct{}
}

type Posts struct {
	svc    *services.PostsService
	events *events.Broker
}

type created struct{ api.Post }

type Session struct {
	store *session.Store
}
