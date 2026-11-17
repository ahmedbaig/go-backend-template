package jobs

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Every    time.Duration
	Jitter   time.Duration
	Timeout  time.Duration
	RunFirst bool
	Run      func(ctx context.Context) error
}

type Runner struct {
	log  *slog.Logger
	jobs []Job
	wg   sync.WaitGroup
	stop context.CancelFunc
}
