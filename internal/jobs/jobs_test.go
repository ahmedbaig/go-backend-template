package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/jobs"
	testkit "github.com/ahmedbaig/go-backend-testing"
)

func TestRunnerRunsAndStops(t *testing.T) {
	var n atomic.Int32
	var sawID atomic.Bool
	r := jobs.NewRunner(slog.New(slog.NewTextHandler(io.Discard, nil)),
		jobs.Job{Name: "tick", Every: 10 * time.Millisecond, RunFirst: true, Run: func(ctx context.Context) error {
			if httpx.CorrelationID(ctx) != "" {
				sawID.Store(true)
			}
			n.Add(1)
			return nil
		}},
		jobs.Job{Name: "flaky", Every: 10 * time.Millisecond, Run: func(context.Context) error { return errors.New("nope") }},
		jobs.Job{Name: "boom", Every: 10 * time.Millisecond, Run: func(context.Context) error { panic("x") }},
	)
	r.Start(context.Background())
	time.Sleep(60 * time.Millisecond)
	testkit.NoError(t, r.Stop(context.Background()))
	after := n.Load()
	time.Sleep(30 * time.Millisecond)
	testkit.True(t, after >= 3, "ran several times")
	testkit.Equal(t, n.Load(), after)
	testkit.True(t, sawID.Load(), "each run gets a correlation id")
}
