package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
)

func Capabilities(c *db.Client, log *slog.Logger, every time.Duration) Job {
	return Job{
		Name:     constants.CapabilitiesJob,
		Every:    every,
		Jitter:   constants.CapabilitiesJitter,
		Timeout:  constants.CapabilitiesTimeout,
		RunFirst: true,
		Run: func(ctx context.Context) error {
			caps, err := c.Capabilities(ctx)
			if err != nil {
				return err
			}
			log.InfoContext(ctx, "db-handler capabilities", slog.Int("databases", len(caps.Databases)), slog.Int("stores", len(caps.Stores)))
			return nil
		},
	}
}
