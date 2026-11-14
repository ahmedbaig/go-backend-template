package services

import (
	"context"
	"errors"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/types"
)

const (
	database = "demo"
	table    = "posts"
)

type PostsService struct {
	posts *db.Service[types.Post]
}

func NewPostsService(c *db.Client) *PostsService {
	return &PostsService{posts: db.NewService[types.Post](c, database, table)}
}

func (s *PostsService) List(ctx context.Context, authorID int64, f types.ListPostsRequest) (types.ListPostsResponse, error) {
	where := db.W().Eq("authorId", authorID)
	if f.Published != nil {
		where = where.Eq("published", *f.Published)
	}
	if f.Search != "" {
		where = where.Field("title", db.Contains(f.Search))
	}
	take := f.Take
	if take == 0 {
		take = 20
	}
	items, err := s.posts.FindMany(ctx, db.Params{
		Where:   where,
		OrderBy: []db.Order{{Field: "createdAt", Direction: db.Desc}},
		Take:    db.Take(take),
		Skip:    f.Skip,
	})
	if err != nil {
		return types.ListPostsResponse{}, upstream(err)
	}
	total, err := s.posts.Count(ctx, where)
	if err != nil {
		return types.ListPostsResponse{}, upstream(err)
	}
	if items == nil {
		items = []types.Post{}
	}
	return types.ListPostsResponse{Items: items, Total: total}, nil
}

func (s *PostsService) Get(ctx context.Context, authorID, id int64) (types.Post, error) {
	p, err := s.posts.FindOne(ctx, db.W().Eq("id", id).Eq("authorId", authorID))
	return p, upstream(err)
}

func (s *PostsService) Create(ctx context.Context, authorID int64, in types.CreatePostRequest) (types.Post, error) {
	p, err := s.posts.Create(ctx, types.Post{AuthorID: authorID, Title: in.Title, Body: in.Body, Published: in.Published})
	return p, upstream(err)
}

func (s *PostsService) Update(ctx context.Context, authorID int64, in types.UpdatePostRequest) (types.Post, error) {
	changes := map[string]any{}
	if in.Title != nil {
		changes["title"] = *in.Title
	}
	if in.Body != nil {
		changes["body"] = *in.Body
	}
	if in.Published != nil {
		changes["published"] = *in.Published
	}
	if len(changes) == 0 {
		return types.Post{}, httpx.E(httpx.CodeValidation).WithMessage("Nothing to update")
	}
	p, err := s.posts.Update(ctx, db.W().Eq("id", in.ID).Eq("authorId", authorID), changes)
	return p, upstream(err)
}

func (s *PostsService) Delete(ctx context.Context, authorID, id int64) error {
	_, err := s.posts.Delete(ctx, db.W().Eq("id", id).Eq("authorId", authorID))
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
