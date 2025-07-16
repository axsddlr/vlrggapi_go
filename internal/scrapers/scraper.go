package scrapers

import (
	"github.com/gofiber/fiber/v2"
)

// Scraper is a generic interface for all scrapers.
type Scraper interface {
	Route() string
	Handler() fiber.Handler
	Description() string
}

// Registry holds all registered scrapers.
var Registry = make([]Scraper, 0)

// RegisterScraper adds a new scraper to the registry.
func RegisterScraper(s Scraper) {
	Registry = append(Registry, s)
}
