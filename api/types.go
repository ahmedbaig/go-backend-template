package api

import (
	"github.com/ahmedbaig/go-backend-types"
)

const (
	PathPosts     = "/api/v1/posts"
	PathPost      = "/api/v1/posts/{id}"
	PathAdminInfo = "/api/v1/admin/info"
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
