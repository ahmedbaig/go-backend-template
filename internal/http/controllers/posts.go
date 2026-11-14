package controllers

import (
	"context"
	"net/http"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
)

type Posts struct {
	svc *services.PostsService
}

func NewPosts(svc *services.PostsService) *Posts { return &Posts{svc: svc} }

type created struct{ api.Post }

func (created) StatusCode() int { return http.StatusCreated }

func (c *Posts) List(ctx context.Context, in api.ListPostsRequest) (api.ListPostsResponse, error) {
	u, err := httpx.User(ctx)
	if err != nil {
		return api.ListPostsResponse{}, err
	}
	return c.svc.List(ctx, u.ID, in)
}

func (c *Posts) Get(ctx context.Context, in api.GetPostRequest) (api.Post, error) {
	u, err := httpx.User(ctx)
	if err != nil {
		return api.Post{}, err
	}
	return c.svc.Get(ctx, u.ID, in.ID)
}

func (c *Posts) Create(ctx context.Context, in api.CreatePostRequest) (created, error) {
	u, err := httpx.User(ctx)
	if err != nil {
		return created{}, err
	}
	p, err := c.svc.Create(ctx, u.ID, in)
	if err == nil {
		httpx.LoggerFrom(ctx).InfoContext(ctx, "post created", "postId", p.ID)
	}
	return created{p}, err
}

func (c *Posts) Update(ctx context.Context, in api.UpdatePostRequest) (api.Post, error) {
	u, err := httpx.User(ctx)
	if err != nil {
		return api.Post{}, err
	}
	return c.svc.Update(ctx, u.ID, in)
}

func (c *Posts) Delete(w http.ResponseWriter, r *http.Request) error {
	in, err := httpx.Bind[api.GetPostRequest](r, httpx.NoBody())
	if err != nil {
		return err
	}
	u, err := httpx.User(r.Context())
	if err != nil {
		return err
	}
	if err := c.svc.Delete(r.Context(), u.ID, in.ID); err != nil {
		return err
	}
	httpx.NoContent(w)
	return nil
}
