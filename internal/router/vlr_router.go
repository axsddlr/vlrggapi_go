package router

import (
	"vlrggapi/internal/scrapers"
	"github.com/gofiber/fiber/v2"
)

func RegisterVlrRoutes(app *fiber.App) {
	vlr := app.Group("/vlr")

	vlr.Get("/news", scrapers.VlrNews)
	vlr.Get("/stats", scrapers.VlrStats)
	vlr.Get("/rankings", scrapers.VlrRankings)
	vlr.Get("/match", scrapers.VlrMatchResults)
	vlr.Get("/live", scrapers.VlrLiveScore)
	vlr.Get("/events", scrapers.VlrEvents)
	vlr.Get("/health", scrapers.Health)
}
