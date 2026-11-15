package controllers

import (
	"context"
	"net/http"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/events"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
)

type Posts struct {
	svc    *services.PostsService
	events *events.Broker
}

func NewPosts(svc *services.PostsService, ev *events.Broker) *Posts {
	return &Posts{svc: svc, events: ev}
}

func (c *Posts) publish(ctx context.Context, userID int64, typ string, data any) {
	c.events.PublishOrLog(ctx, events.UserChannel(api.TopicPosts, userID), typ, data)
}

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
	if err != nil {
		return created{}, err
	}
	httpx.LoggerFrom(ctx).InfoContext(ctx, "post created", "postId", p.ID)
	c.publish(ctx, u.ID, api.EventPostCreated, p)
	return created{p}, nil
}

func (c *Posts) Update(ctx context.Context, in api.UpdatePostRequest) (api.Post, error) {
	u, err := httpx.User(ctx)
	if err != nil {
		return api.Post{}, err
	}
	p, err := c.svc.Update(ctx, u.ID, in)
	if err != nil {
		return api.Post{}, err
	}
	c.publish(ctx, u.ID, api.EventPostUpdated, p)
	return p, nil
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
	c.publish(r.Context(), u.ID, api.EventPostDeleted, api.PostDeleted{ID: in.ID})
	httpx.NoContent(w)
	return nil
}
