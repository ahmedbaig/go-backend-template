package services

import (
	"context"
	"errors"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-db-handler/types/demo"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-types"
)

type PostsService struct {
	posts *db.Service[demo.Post]
}

func NewPostsService(c *db.Client) *PostsService {
	return &PostsService{posts: demo.PostService(c)}
}

var f = demo.PostFields

func toAPI(p demo.Post) api.Post {
	return api.Post{
		ID: p.ID, AuthorID: p.AuthorID, Title: p.Title, Body: p.Body, Views: p.Views, Published: p.Published,
		Timestamps: types.Timestamps{CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt},
	}
}

func (s *PostsService) List(ctx context.Context, authorID int64, in api.ListPostsRequest) (api.ListPostsResponse, error) {
	where := db.W().Eq(f.AuthorID, authorID)
	if in.Published != nil {
		where = where.Eq(f.Published, *in.Published)
	}
	if in.Search != "" {
		where = where.Field(f.Title, db.Contains(in.Search))
	}
	items, err := s.posts.FindMany(ctx, db.Params{
		Where:   where,
		OrderBy: []db.Order{{Field: f.CreatedAt, Direction: db.Desc}},
		Take:    db.Take(in.Limit()),
		Skip:    in.Offset(),
	})
	if err != nil {
		return api.ListPostsResponse{}, upstream(err)
	}
	total, err := s.posts.Count(ctx, where)
	if err != nil {
		return api.ListPostsResponse{}, upstream(err)
	}
	return types.MapPage(types.NewPage(items, total, in.PageRequest), toAPI), nil
}

func (s *PostsService) Get(ctx context.Context, authorID, id int64) (api.Post, error) {
	p, err := s.posts.FindOne(ctx, db.W().Eq(f.ID, id).Eq(f.AuthorID, authorID))
	return toAPI(p), upstream(err)
}

func (s *PostsService) Create(ctx context.Context, authorID int64, in api.CreatePostRequest) (api.Post, error) {
	p, err := s.posts.Create(ctx, demo.Post{AuthorID: authorID, Title: in.Title, Body: in.Body, Published: in.Published})
	return toAPI(p), upstream(err)
}

func (s *PostsService) Update(ctx context.Context, authorID int64, in api.UpdatePostRequest) (api.Post, error) {
	changes := map[string]any{}
	if in.Title != nil {
		changes[f.Title] = *in.Title
	}
	if in.Body != nil {
		changes[f.Body] = *in.Body
	}
	if in.Published != nil {
		changes[f.Published] = *in.Published
	}
	if len(changes) == 0 {
		return api.Post{}, httpx.E(httpx.CodeValidation).WithMessage("Nothing to update")
	}
	p, err := s.posts.Update(ctx, db.W().Eq(f.ID, in.ID).Eq(f.AuthorID, authorID), changes)
	return toAPI(p), upstream(err)
}

func (s *PostsService) Delete(ctx context.Context, authorID, id int64) error {
	_, err := s.posts.Delete(ctx, db.W().Eq(f.ID, id).Eq(f.AuthorID, authorID))
	return upstream(err)
}

func upstream(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, db.ErrNotFound):
		return httpx.E(httpx.CodeNotFound).WithMessage("Post not found").Wrap(err)
	case errors.Is(err, db.ErrConflict):
		return httpx.E(httpx.CodeConflict).Wrap(err)
	}
	return httpx.E(httpx.CodeBadGateway).Wrap(err)
}
