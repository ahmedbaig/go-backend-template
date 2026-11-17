package api

import (
	"github.com/ahmedbaig/go-backend-types"
	"github.com/ahmedbaig/go-backend-types/client"
)

type Post struct {
	ID        int64  `json:"id"`
	AuthorID  int64  `json:"authorId"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Views     int    `json:"views"`
	Published bool   `json:"published"`
	types.Timestamps
}

type ListPostsRequest struct {
	types.PageRequest
	Published *bool  `query:"published"`
	Search    string `query:"q" validate:"omitempty,max=100"`
}

type ListPostsResponse = types.Page[Post]

type GetPostRequest = types.IDRequest

type CreatePostRequest struct {
	Title     string `json:"title" validate:"required,min=3,max=255"`
	Body      string `json:"body" validate:"max=20000"`
	Published bool   `json:"published"`
}

type UpdatePostRequest struct {
	ID        int64   `path:"id" json:"-" validate:"required,min=1"`
	Title     *string `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Body      *string `json:"body,omitempty" validate:"omitempty,max=20000"`
	Published *bool   `json:"published,omitempty"`
}

type InfoResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
}

type Client struct {
	c *client.Client
}

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

type PostDeleted struct {
	ID int64 `json:"id"`
}
