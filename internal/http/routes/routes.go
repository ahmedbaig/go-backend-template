package routes

import (
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/http/controllers"
	"github.com/ahmedbaig/go-backend-template/internal/http/services"
	"github.com/ahmedbaig/go-backend-template/internal/session"
)

func Register(app *httpx.App, ctx *appctx.Context) {
	posts := controllers.NewPosts(services.NewPostsService(ctx.DB), ctx.Events)
	sess := controllers.NewSession(ctx.Sessions)
	live := controllers.NewEvents(ctx.Events, ctx.Topics, ctx.Config.Events.Heartbeat.Std(), app.Draining())

	root := app.Group("/api/v1", ctx.Sessions.Load())
	root.Get("/session", httpx.HandlerFunc(sess.Get), session.Require())
	root.Delete("/session", httpx.HandlerFunc(sess.Logout))
	if ctx.Config.DevLoginEnabled() {
		root.Post("/session/dev-login", httpx.HandlerFunc(sess.DevLogin))
	}

	v1 := root.Group("", session.Require())
	v1.Get("/events", httpx.HandlerFunc(live.Stream))

	read := httpx.RequireFeature(api.ScopePostsRead)
	write := httpx.RequireFeature(api.ScopePostsWrite)
	p := v1.Group("/posts")
	p.Get("", httpx.Handle(posts.List, httpx.NoBody()), read)
	p.Post("", httpx.Handle(posts.Create), write)
	p.Get("/{id}", httpx.Handle(posts.Get, httpx.NoBody()), read)
	p.Patch("/{id}", httpx.Handle(posts.Update), write)
	p.Delete("/{id}", httpx.HandlerFunc(posts.Delete), write)

	admin := v1.Group("/admin", httpx.RequireRole("admin"))
	admin.Get("/info", httpx.HandlerFunc(info(ctx)))
}
