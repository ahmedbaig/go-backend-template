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

	root := app.Group("", ctx.Sessions.Load())
	root.Get(api.PathSession, httpx.HandlerFunc(sess.Get), session.Require())
	root.Delete(api.PathSession, httpx.HandlerFunc(sess.Logout))
	if ctx.Config.DevLoginEnabled() {
		root.Post(api.PathSessionDevLogin, httpx.HandlerFunc(sess.DevLogin))
	}

	v1 := root.Group("", session.Require())
	v1.Get(api.PathEvents, httpx.HandlerFunc(live.Stream))

	read := httpx.RequireFeature(api.ScopePostsRead)
	write := httpx.RequireFeature(api.ScopePostsWrite)
	v1.Get(api.PathPosts, httpx.Handle(posts.List, httpx.NoBody()), read)
	v1.Post(api.PathPosts, httpx.Handle(posts.Create), write)
	v1.Get(api.PathPost, httpx.Handle(posts.Get, httpx.NoBody()), read)
	v1.Patch(api.PathPost, httpx.Handle(posts.Update), write)
	v1.Delete(api.PathPost, httpx.HandlerFunc(posts.Delete), write)

	admin := v1.Group("", httpx.RequireRole(api.RoleAdmin))
	admin.Get(api.PathAdminInfo, httpx.HandlerFunc(info(ctx)))
}
