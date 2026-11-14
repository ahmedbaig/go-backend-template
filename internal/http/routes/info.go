package routes

import (
	"net/http"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-template/internal/appctx"
	"github.com/ahmedbaig/go-backend-template/internal/config"
)

func info(ctx *appctx.Context) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		httpx.OK(w, map[string]any{
			"name":        config.Name,
			"version":     ctx.Version,
			"environment": ctx.Config.Environment,
			"status":      ctx.Prober.Status(),
		})
		return nil
	}
}
