package controllers

import (
	"context"
	"net/http"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/http/middleware"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
	"github.com/ahmedbaig/go-backend-template/internal/types"
)

type Posts struct {
	svc *services.PostsService
}

func NewPosts(svc *services.PostsService) *Posts { return &Posts{svc: svc} }

type created struct{ types.Post }

func (created) StatusCode() int { return http.StatusCreated }

func userID(ctx context.Context) (int64, error) {
	u, ok := middleware.UserFrom(ctx)
	if !ok {
		return 0, httpx.E(httpx.CodeUnauthenticated)
	}
	return u.ID, nil
}

func (c *Posts) List(ctx context.Context, in types.ListPostsRequest) (types.ListPostsResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return types.ListPostsResponse{}, err
	}
	return c.svc.List(ctx, uid, in)
}

func (c *Posts) Get(ctx context.Context, in types.GetPostRequest) (types.Post, error) {
	uid, err := userID(ctx)
	if err != nil {
		return types.Post{}, err
	}
	return c.svc.Get(ctx, uid, in.ID)
}

func (c *Posts) Create(ctx context.Context, in types.CreatePostRequest) (created, error) {
	uid, err := userID(ctx)
	if err != nil {
		return created{}, err
	}
	p, err := c.svc.Create(ctx, uid, in)
	if err == nil {
		httpx.LoggerFrom(ctx).InfoContext(ctx, "post created", "postId", p.ID)
	}
	return created{p}, err
}

func (c *Posts) Update(ctx context.Context, in types.UpdatePostRequest) (types.Post, error) {
	uid, err := userID(ctx)
	if err != nil {
		return types.Post{}, err
	}
	return c.svc.Update(ctx, uid, in)
}

func (c *Posts) Delete(w http.ResponseWriter, r *http.Request) error {
	in, err := httpx.Bind[types.GetPostRequest](r, httpx.NoBody())
	if err != nil {
		return err
	}
	uid, err := userID(r.Context())
	if err != nil {
		return err
	}
	if err := c.svc.Delete(r.Context(), uid, in.ID); err != nil {
		return err
	}
	httpx.NoContent(w)
	return nil
}
