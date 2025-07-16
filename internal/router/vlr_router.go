package router

import (
	"vlrggapi/internal/scrapers"
	"github.com/gofiber/fiber/v2"
)

func RegisterVlrRoutes(app *fiber.App) {
	vlr := app.Group("/vlr")

	// Register all modular scrapers
	for _, s := range scrapers.Registry {
		vlr.Get(s.Route(), s.Handler())
	}

	// Legacy/manual endpoints (for backward compatibility or not yet modularized)
	vlr.Get("/news", scrapers.VlrNews)
	vlr.Get("/stats", scrapers.VlrStats)
	vlr.Get("/rankings", scrapers.VlrRankings)
	vlr.Get("/match", scrapers.VlrMatchResults)
	vlr.Get("/live", scrapers.VlrLiveScore)
	vlr.Get("/events", scrapers.VlrEvents)
	vlr.Get("/health", scrapers.Health)
}
