package e2e

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-http"
	dbv1 "github.com/ahmedbaig/go-backend-proto/gen/go/ahmedbaig/dbhandler/v1"
	"github.com/ahmedbaig/go-backend-proto/grpcx"
	"github.com/ahmedbaig/go-backend-template/api"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/ahmedbaig/go-backend-template/internal/server"
	testkit "github.com/ahmedbaig/go-backend-testing"
	"github.com/ahmedbaig/go-backend-types"
	"github.com/ahmedbaig/go-backend-types/client"
	"github.com/alicebob/miniredis/v2"
)

var everything = []string{api.ScopePostsRead, api.ScopePostsWrite}

func baseConfig(t *testing.T) config.App {
	t.Helper()
	t.Setenv("CI", "true")
	c := config.Defaults()
	c.Environment = "testing"
	c.Redis.Addr = miniredis.RunT(t).Addr()
	c.Session.DevLogin = true
	c.Session.Secure = false
	c.Events.Heartbeat = 0
	c.Jobs.CapabilitiesEvery = 0
	return c
}

func setup(t *testing.T) (*testkit.Client, *testkit.Stub, *testkit.LogCapture, http.Handler) {
	t.Helper()
	dbh := testkit.NewStub(t)
	c := baseConfig(t)
	c.DBHandler.Transport = "http"
	c.DBHandler.URL = dbh.URL()
	log, logs := testkit.Logger(t)
	app, err := server.Build(c, log)
	testkit.NoError(t, err)
	return testkit.NewClient(t, app.Handler()), dbh, logs, app.Handler()
}

func login(t *testing.T, c *testkit.Client, roles []string, features ...string) *testkit.Client {
	t.Helper()
	authed, _ := loginCookie(t, c, roles, features...)
	return authed
}

func loginCookie(t *testing.T, c *testkit.Client, roles []string, features ...string) (*testkit.Client, string) {
	t.Helper()
	res := c.Post(api.PathSessionDevLogin, api.DevLoginRequest{ID: 7, Email: "alice@example.com", Name: "Alice", Roles: roles, Features: features}).Expect(200)
	cookie := res.Header.Get("Set-Cookie")
	testkit.True(t, strings.Contains(cookie, "HttpOnly") && strings.Contains(cookie, "SameSite=Lax"), "session cookie flags")
	kv := strings.SplitN(cookie, ";", 2)[0]
	return c.WithHeader("Cookie", kv), kv
}

func TestSessionLifecycle(t *testing.T) {
	anon, _, _, _ := setup(t)
	anon.Get(api.PathSession).ExpectError(401, httpx.CodeUnauthenticated)
	anon.Post(api.PathSessionDevLogin, map[string]any{"email": "nope"}).ExpectField("email", "email")

	c := login(t, anon, nil, everything...)
	s := testkit.Decode[api.SessionResponse](c.Get(api.PathSession).Expect(200))
	testkit.Equal(t, s.User.ID, int64(7))
	testkit.True(t, s.User.Can(api.ScopePostsWrite), "scopes kept")

	c.Delete(api.PathSession).Expect(204)
	c.Get(api.PathSession).ExpectError(401, httpx.CodeSessionExpired)
}

func TestDevLoginOffOutsideLocal(t *testing.T) {
	c := baseConfig(t)
	c.Environment = "production"
	c.DBHandler.Transport = "http"
	c.DBHandler.URL = "http://127.0.0.1:1"
	log, _ := testkit.Logger(t)
	app, err := server.Build(c, log)
	testkit.NoError(t, err)
	testkit.NewClient(t, app.Handler()).Post(api.PathSessionDevLogin, api.DevLoginRequest{Email: "a@example.com", Name: "A"}).ExpectError(404, httpx.CodeRouteNotFound)
}

func TestScopes(t *testing.T) {
	anon, _, _, _ := setup(t)
	reader := login(t, anon, nil, api.ScopePostsRead)
	reader.Post("/api/v1/posts", map[string]any{"title": "Nope"}).ExpectError(403, httpx.CodeForbidden)
	nobody := login(t, anon, nil)
	nobody.Get("/api/v1/posts").ExpectError(403, httpx.CodeForbidden)
	nobody.Get(api.PathEvents+"?topics=posts").ExpectError(403, httpx.CodeForbidden)
	nobody.Get("/api/v1/admin/info").ExpectError(403, httpx.CodeForbidden)
	admin := login(t, anon, []string{"admin"})
	info := testkit.Decode[map[string]any](admin.Get("/api/v1/admin/info").Expect(200))
	testkit.Equal(t, info["name"], any(constants.Name))
}

func TestListPostsScopedToUser(t *testing.T) {
	anon, dbh, _, _ := setup(t)
	c := login(t, anon, nil, everything...)
	dbh.Handle("POST "+db.PathQuery, func(w http.ResponseWriter, r *http.Request) {
		var q db.Query
		_ = json.NewDecoder(r.Body).Decode(&q)
		if q.Method == db.Count {
			_, _ = w.Write([]byte(`{"count":1}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":1,"authorId":7,"title":"Hello","published":true}]}`))
	})

	res := testkit.Decode[api.ListPostsResponse](c.Get("/api/v1/posts?published=true&q=hel&take=5").Expect(200))
	testkit.Equal(t, res.Total, int64(1))
	testkit.Equal(t, res.Items[0].Title, "Hello")

	calls := dbh.RequireCalled("POST "+db.PathQuery, 2)
	var q db.Query
	calls[0].JSON(t, &q)
	testkit.Equal(t, q.Service, "posts")
	testkit.Equal(t, q.Params.Where.Fields["authorId"][db.OpEq], any(float64(7)))
	testkit.Equal(t, q.Params.Where.Fields["title"][db.OpContains], any("hel"))
	testkit.Equal(t, *q.Params.Take, 5)
	testkit.True(t, calls[0].Header.Get(httpx.HeaderCorrelationID) != "", "correlation id forwarded to db-handler")
}

func TestCreatePost(t *testing.T) {
	anon, dbh, logs, _ := setup(t)
	c := login(t, anon, nil, everything...)
	dbh.On("POST "+db.PathQuery, 200, `{"data":{"id":9,"authorId":7,"title":"New post"}}`)
	p := testkit.Decode[api.Post](c.Post("/api/v1/posts", map[string]any{"title": "New post"}).Expect(201))
	testkit.Equal(t, p.ID, int64(9))
	e := logs.RequireLogged(t, "post created")
	testkit.Equal(t, e.Attrs["userId"], any(int64(7)))
	c.Post("/api/v1/posts", map[string]any{"title": "x"}).ExpectField("title", "min")
}

func TestUpstreamErrorsMapped(t *testing.T) {
	anon, dbh, _, _ := setup(t)
	c := login(t, anon, nil, everything...)
	dbh.On("POST "+db.PathQuery, 404, `{"errors":[{"code":"E404-1","message":"not found"}]}`)
	c.Get("/api/v1/posts/3").ExpectError(404, httpx.CodeNotFound)
	c.Delete("/api/v1/posts/abc").ExpectError(400, httpx.CodeValidation)
}

func TestLiveEventsOverSSE(t *testing.T) {
	anon, dbh, _, h := setup(t)
	res, cookie := loginCookie(t, anon, nil, everything...)
	dbh.On("POST "+db.PathQuery, 200, `{"data":{"id":11,"authorId":7,"title":"Live one"}}`)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+api.PathEvents+"?topics=posts", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cookie", cookie)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := http.DefaultClient.Do(req.WithContext(ctx))
	testkit.NoError(t, err)
	defer stream.Body.Close()
	testkit.Equal(t, stream.Header.Get("Content-Type"), "text/event-stream")

	res.WithHeader(httpx.HeaderCorrelationID, "corr-live").Post("/api/v1/posts", map[string]any{"title": "Live one"}).Expect(201)

	sc := bufio.NewScanner(stream.Body)
	var got []string
	for sc.Scan() {
		got = append(got, sc.Text())
		if strings.HasPrefix(sc.Text(), "data: ") {
			break
		}
	}
	all := strings.Join(got, "\n")
	testkit.True(t, strings.Contains(all, "event: posts.created"), "event type: "+all)
	testkit.True(t, strings.Contains(all, `"title":"Live one"`), "event data: "+all)
	testkit.True(t, strings.Contains(all, `"correlationId":"corr-live"`), "correlation id carried into the event: "+all)
}

type fakeDB struct {
	dbv1.UnimplementedDBHandlerServiceServer
	ids chan [2]string
}

func (f *fakeDB) Query(ctx context.Context, q *dbv1.QueryRequest) (*dbv1.QueryResponse, error) {
	u, _ := types.UserFrom(ctx)
	f.ids <- [2]string{httpx.CorrelationID(ctx), u.Email}
	if db.Method(q.GetMethod()) == db.Count {
		n := int64(1)
		return &dbv1.QueryResponse{Count: &n}, nil
	}
	if db.Method(q.GetMethod()) == db.FindOne {
		return nil, httpx.E(httpx.CodeNotFound)
	}
	return &dbv1.QueryResponse{DataJson: []byte(`[{"id":1,"authorId":7,"title":"Over gRPC"}]`)}, nil
}

func TestDBHandlerOverGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	testkit.NoError(t, err)
	f := &fakeDB{ids: make(chan [2]string, 4)}
	gs := grpcx.NewServer(grpcx.ServerOptions{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), ServiceTokens: []string{"svc-token"}})
	dbv1.RegisterDBHandlerServiceServer(gs, f)
	gs.MarkServing(dbv1.DBHandlerService_ServiceDesc.ServiceName)
	go func() { _ = gs.Serve(lis) }()
	defer gs.Stop()

	cfg := baseConfig(t)
	cfg.DBHandler.Transport = "grpc"
	cfg.DBHandler.GRPCAddr = lis.Addr().String()
	cfg.DBHandler.APIKey = "svc-token"
	log, _ := testkit.Logger(t)
	app, err := server.Build(cfg, log)
	testkit.NoError(t, err)
	c := login(t, testkit.NewClient(t, app.Handler()), nil, everything...).WithHeader(httpx.HeaderCorrelationID, "corr-grpc")

	res := testkit.Decode[api.ListPostsResponse](c.Get("/api/v1/posts").Expect(200))
	testkit.Equal(t, res.Items[0].Title, "Over gRPC")
	got := <-f.ids
	testkit.Equal(t, got[0], "corr-grpc")
	testkit.Equal(t, got[1], "alice@example.com")

	c.Get("/api/v1/posts/5").ExpectError(404, httpx.CodeNotFound)
}

func TestHealth(t *testing.T) {
	c, _, _, _ := setup(t)
	c.Get("/healthz").Expect(200)
	c.Get("/healthp").Expect(200)
}

func TestAPIClientContract(t *testing.T) {
	anon, dbh, _, h := setup(t)
	dbh.On("POST "+db.PathQuery, 200, `{"data":{"id":4,"authorId":7,"title":"Via client"}}`)
	srv := httptest.NewServer(h)
	defer srv.Close()
	_, cookie := loginCookie(t, anon, []string{"admin"}, everything...)

	cl := api.NewClient(srv.URL, client.WithHeader("Cookie", cookie))
	p, err := cl.CreatePost(context.Background(), api.CreatePostRequest{Title: "Via client"})
	testkit.NoError(t, err)
	testkit.Equal(t, p.ID, int64(4))

	_, err = cl.CreatePost(context.Background(), api.CreatePostRequest{Title: "x"})
	var ae *types.APIError
	testkit.True(t, errors.As(err, &ae) && ae.Invalid() && ae.Code() == string(httpx.CodeValidation), "validation error decoded")

	s, err := cl.Session(context.Background())
	testkit.NoError(t, err)
	testkit.Equal(t, s.User.ID, int64(7))
}
