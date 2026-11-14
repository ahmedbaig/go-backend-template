package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/types"
	testkit "github.com/ahmedbaig/go-backend-testing"
)

const alice = `{"id":7,"email":"alice@example.com","name":"Alice","roles":["admin"]}`

func setup(t *testing.T) (*testkit.Client, *testkit.Stub, *testkit.LogCapture) {
	t.Helper()
	t.Setenv("CI", "true")
	dbh := testkit.NewStub(t)
	c := config.Defaults()
	c.Environment = "testing"
	c.DBHandler.URL = dbh.URL()
	log, logs := testkit.Logger(t)
	app, err := build(c, log)
	testkit.NoError(t, err)
	return testkit.NewClient(t, app.Handler()).WithHeader("X-User", alice), dbh, logs
}

func TestListPostsScopedToUser(t *testing.T) {
	c, dbh, _ := setup(t)
	dbh.Handle("POST "+db.PathQuery, func(w http.ResponseWriter, r *http.Request) {
		var q db.Query
		_ = json.NewDecoder(r.Body).Decode(&q)
		if q.Method == db.Count {
			_, _ = w.Write([]byte(`{"count":1}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":1,"authorId":7,"title":"Hello","published":true}]}`))
	})

	res := testkit.Decode[types.ListPostsResponse](c.Get("/api/v1/posts?published=true&q=hel&take=5").Expect(200))
	testkit.Equal(t, res.Total, int64(1))
	testkit.Equal(t, res.Items[0].Title, "Hello")

	calls := dbh.RequireCalled("POST "+db.PathQuery, 2)
	var q db.Query
	calls[0].JSON(t, &q)
	testkit.Equal(t, q.Database, "demo")
	testkit.Equal(t, q.Service, "posts")
	testkit.Equal(t, q.Params.Where.Fields["authorId"][db.OpEq], any(float64(7)))
	testkit.Equal(t, q.Params.Where.Fields["title"][db.OpContains], any("hel"))
	testkit.Equal(t, *q.Params.Take, 5)
	testkit.True(t, calls[0].Header.Get(httpx.HeaderCorrelationID) != "", "correlation id forwarded to db-handler")
}

func TestCreatePost(t *testing.T) {
	c, dbh, logs := setup(t)
	dbh.On("POST "+db.PathQuery, 200, `{"data":{"id":9,"authorId":7,"title":"New post"}}`)
	p := testkit.Decode[types.Post](c.Post("/api/v1/posts", map[string]any{"title": "New post"}).Expect(201))
	testkit.Equal(t, p.ID, int64(9))
	e := logs.RequireLogged(t, "post created")
	testkit.Equal(t, e.Attrs["userId"], any(int64(7)))

	c.Post("/api/v1/posts", map[string]any{"title": "x"}).ExpectError(400, httpx.CodeValidation)
	c.Post("/api/v1/posts", map[string]any{"title": "x"}).ExpectField("title", "min")
}

func TestUpstreamErrorsMapped(t *testing.T) {
	c, dbh, _ := setup(t)
	dbh.On("POST "+db.PathQuery, 404, `{"errors":[{"code":"E404-1","message":"not found"}]}`)
	c.Get("/api/v1/posts/3").ExpectError(404, httpx.CodeNotFound)
	c.Delete("/api/v1/posts/abc").ExpectError(400, httpx.CodeValidation)
}

func TestAuth(t *testing.T) {
	c, _, _ := setup(t)
	c.WithHeader("X-User", "").Get("/api/v1/posts").ExpectError(401, httpx.CodeUnauthenticated)
	c.WithHeader("X-User", `{"id":1,"roles":[]}`).Get("/api/v1/admin/info").ExpectError(403, httpx.CodeForbidden)
	info := testkit.Decode[map[string]any](c.Get("/api/v1/admin/info").Expect(200))
	testkit.Equal(t, info["name"], any(config.Name))
}

func TestHealth(t *testing.T) {
	c, _, _ := setup(t)
	c.Get("/healthz").Expect(200)
	c.Get("/healthp").Expect(200)
}
