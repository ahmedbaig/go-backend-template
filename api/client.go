package api

import (
	"context"

	"github.com/ahmedbaig/go-backend-types/client"
)

type Client struct {
	c *client.Client
}

func NewClient(baseURL string, opts ...client.Option) *Client {
	return &Client{c: client.New(baseURL, opts...)}
}

func (c *Client) ListPosts(ctx context.Context, req ListPostsRequest) (ListPostsResponse, error) {
	return client.Get[ListPostsResponse](ctx, c.c, PathPosts, req)
}

func (c *Client) GetPost(ctx context.Context, id int64) (Post, error) {
	return client.Get[Post](ctx, c.c, client.Path(PathPost, id), nil)
}

func (c *Client) CreatePost(ctx context.Context, req CreatePostRequest) (Post, error) {
	return client.Post[Post](ctx, c.c, PathPosts, req)
}

func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest) (Post, error) {
	return client.Patch[Post](ctx, c.c, client.Path(PathPost, req.ID), req)
}

func (c *Client) DeletePost(ctx context.Context, id int64) error {
	return client.Delete(ctx, c.c, client.Path(PathPost, id))
}

func (c *Client) AdminInfo(ctx context.Context) (InfoResponse, error) {
	return client.Get[InfoResponse](ctx, c.c, PathAdminInfo, nil)
}

func (c *Client) Session(ctx context.Context) (SessionResponse, error) {
	return client.Get[SessionResponse](ctx, c.c, PathSession, nil)
}
