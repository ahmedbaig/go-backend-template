package events_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-http/sse"
	"github.com/ahmedbaig/go-backend-template/internal/events"
	testkit "github.com/ahmedbaig/go-backend-testing"
	"github.com/ahmedbaig/go-backend-types"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func broker(t *testing.T) *events.Broker {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return events.NewBroker(rdb, events.Options{Prefix: "t:"})
}

func next(t *testing.T, ch <-chan sse.Event) sse.Event {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(2 * time.Second):
		t.Fatal("no event")
	}
	return sse.Event{}
}

func TestLiveAndReplay(t *testing.T) {
	b := broker(t)
	ctx, cancel := context.WithCancel(httpx.WithIDs(context.Background(), "r1", "c1"))
	defer cancel()

	first, err := b.Publish(ctx, "posts:7", "posts.created", map[string]any{"id": 1})
	testkit.NoError(t, err)
	_, err = b.Publish(ctx, "posts:7", "posts.created", map[string]any{"id": 2})
	testkit.NoError(t, err)

	ch, err := b.Subscribe(ctx, []string{"posts:7"}, first)
	testkit.NoError(t, err)
	replayed := next(t, ch)
	testkit.Equal(t, replayed.Type, "posts.created")
	testkit.True(t, strings.Contains(string(replayed.Data), `"id":2`), "replays only events after Last-Event-ID")
	testkit.True(t, strings.Contains(string(replayed.Data), `"correlationId":"c1"`), "keeps correlation id")

	_, err = b.Publish(ctx, "posts:7", "posts.deleted", map[string]any{"id": 2})
	testkit.NoError(t, err)
	live := next(t, ch)
	testkit.Equal(t, live.Type, "posts.deleted")
	testkit.True(t, events.After(live.ID, replayed.ID), "ids increase")
}

func TestOtherUsersChannelIsSeparate(t *testing.T) {
	b := broker(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := b.Subscribe(ctx, []string{"posts:7"}, "")
	testkit.NoError(t, err)
	_, _ = b.Publish(ctx, "posts:8", "posts.created", map[string]any{"id": 9})
	_, _ = b.Publish(ctx, "posts:7", "posts.created", map[string]any{"id": 3})
	e := next(t, ch)
	testkit.True(t, strings.Contains(string(e.Data), `"id":3`), "only own channel")
}

func TestTopicsResolve(t *testing.T) {
	ts := events.NewTopics(events.Topic{Name: "posts", Scope: "posts:read", PerUser: true}, events.Topic{Name: "alerts"})
	u := types.User{ID: 7, Features: []string{"posts:read"}}
	chans, err := ts.Resolve(u, "posts, alerts,posts")
	testkit.NoError(t, err)
	testkit.Equal(t, strings.Join(chans, ","), "posts:7,alerts")

	_, err = ts.Resolve(types.User{ID: 1}, "posts")
	testkit.ErrorIs(t, err, httpx.E(httpx.CodeForbidden))
	_, err = ts.Resolve(u, "nope")
	testkit.ErrorIs(t, err, httpx.E(httpx.CodeValidation))
	_, err = ts.Resolve(u, "")
	testkit.ErrorIs(t, err, httpx.E(httpx.CodeValidation))
}

func TestAfter(t *testing.T) {
	testkit.True(t, events.After("2-0", "1-5"), "ms wins")
	testkit.True(t, events.After("1-6", "1-5"), "seq wins")
	testkit.True(t, !events.After("1-5", "1-5"), "equal is not after")
}
