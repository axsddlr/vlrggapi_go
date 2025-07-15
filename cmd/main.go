package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"vlrggapi/internal/router"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "vlrggapi",
		ServerHeader: "vlrggapi",
	})

	app.Use(logger.New())

	// Global rate limiting (600 requests/minute per IP)
	app.Use(limiter.New(limiter.Config{
		Max:        600,
		Expiration: 60 * 1000 * 1000 * 1000, // 1 minute in nanoseconds
	}))

	// Register VLR router
	router.RegisterVlrRoutes(app)

	// Root redirect to docs
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs", fiber.StatusFound)
	})

	// Simple /docs endpoint
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to vlrggapi! Documentation coming soon.",
			"endpoints": []string{
				"/vlr/news",
				"/vlr/stats",
				"/vlr/rankings",
				"/vlr/match",
				"/vlr/health",
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
