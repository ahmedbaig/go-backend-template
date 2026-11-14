package routes

import (
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/http/controllers"
	"github.com/ahmedbaig/go-backend-template/internal/http/middleware"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
)

func Register(app *httpx.App, ctx *appctx.Context) {
	posts := controllers.NewPosts(services.NewPostsService(ctx.DB))

	v1 := app.Group("/api/v1", middleware.RequireUser())

	p := v1.Group("/posts")
	p.Get("", httpx.Handle(posts.List, httpx.NoBody()))
	p.Post("", httpx.Handle(posts.Create))
	p.Get("/{id}", httpx.Handle(posts.Get, httpx.NoBody()))
	p.Patch("/{id}", httpx.Handle(posts.Update))
	p.Delete("/{id}", httpx.HandlerFunc(posts.Delete))

	admin := v1.Group("/admin", middleware.RequireRole("admin"))
	admin.Get("/info", httpx.HandlerFunc(info(ctx)))
}
