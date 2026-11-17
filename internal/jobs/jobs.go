package jobs

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-logging"
	"github.com/google/uuid"
)

func NewRunner(log *slog.Logger, jobs ...Job) *Runner {
	return &Runner{log: log, jobs: jobs}
}

func (r *Runner) Start(ctx context.Context) {
	ctx, r.stop = context.WithCancel(context.WithoutCancel(ctx))
	for _, j := range r.jobs {
		if j.Every <= 0 {
			continue
		}
		r.wg.Add(1)
		go r.loop(ctx, j)
	}
}

func (r *Runner) Stop(context.Context) error {
	if r.stop != nil {
		r.stop()
	}
	r.wg.Wait()
	return nil
}

func (r *Runner) loop(ctx context.Context, j Job) {
	defer r.wg.Done()
	if j.RunFirst {
		r.once(ctx, j)
	}
	for {
		wait := j.Every
		if j.Jitter > 0 {
			wait += rand.N(j.Jitter)
		}
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			r.once(ctx, j)
		}
	}
}

func (r *Runner) once(ctx context.Context, j Job) {
	id := uuid.NewString()
	ctx = httpx.WithIDs(ctx, id, id)
	ctx = logging.WithAttrs(ctx, slog.String("job", j.Name), slog.String("requestId", id), slog.String("correlationId", id))
	ctx = httpx.WithLogger(ctx, r.log)
	if j.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, j.Timeout)
		defer cancel()
	}
	start := time.Now()
	defer func() {
		if p := recover(); p != nil {
			r.log.ErrorContext(ctx, "job panicked", slog.Any("panic", p))
		}
	}()
	if err := j.Run(ctx); err != nil {
		r.log.WarnContext(ctx, "job failed", slog.Duration("took", time.Since(start)), slog.Any("error", err))
		return
	}
	r.log.DebugContext(ctx, "job done", slog.Duration("took", time.Since(start)))
}
