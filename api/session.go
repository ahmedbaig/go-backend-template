package api

import "github.com/ahmedbaig/go-backend-types"

const (
	PathSession         = "/api/v1/session"
	PathSessionDevLogin = "/api/v1/session/dev-login"
	PathEvents          = "/api/v1/events"
)

const (
	ScopePostsRead  = "posts:read"
	ScopePostsWrite = "posts:write"
)

type SessionResponse struct {
	User types.User `json:"user"`
}

type DevLoginRequest struct {
	ID       int64    `json:"id" validate:"omitempty,min=1"`
	Email    string   `json:"email" validate:"required,email"`
	Name     string   `json:"name" validate:"required,max=100"`
	Roles    []string `json:"roles" validate:"max=20,dive,max=50"`
	Features []string `json:"features" validate:"max=50,dive,max=50"`
}

const (
	TopicPosts       = "posts"
	EventPostCreated = "posts.created"
	EventPostUpdated = "posts.updated"
	EventPostDeleted = "posts.deleted"
)

type PostDeleted struct {
	ID int64 `json:"id"`
}
