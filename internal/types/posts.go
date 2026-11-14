package types

import "time"

type Post struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"authorId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Views     int       `json:"views"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListPostsRequest struct {
	Published *bool  `query:"published"`
	Search    string `query:"q" validate:"omitempty,max=100"`
	Take      int    `query:"take" validate:"omitempty,min=1,max=100"`
	Skip      int    `query:"skip" validate:"omitempty,min=0"`
}

type ListPostsResponse struct {
	Items []Post `json:"items"`
	Total int64  `json:"total"`
}

type GetPostRequest struct {
	ID int64 `path:"id" validate:"required,min=1"`
}

type CreatePostRequest struct {
	Title     string `json:"title" validate:"required,min=3,max=255"`
	Body      string `json:"body" validate:"max=20000"`
	Published bool   `json:"published"`
}

type UpdatePostRequest struct {
	ID        int64   `path:"id" validate:"required,min=1"`
	Title     *string `json:"title" validate:"omitempty,min=3,max=255"`
	Body      *string `json:"body" validate:"omitempty,max=20000"`
	Published *bool   `json:"published"`
}
