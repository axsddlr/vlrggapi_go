package main

// @title vlrggapi
// @version 1.0
// @description REST API for scraping Valorant esports data from vlr.gg
// @contact.name vlrggapi Maintainers
// @contact.url https://github.com/yourusername/vlrggapi
// @license.name MIT
// @host localhost:3001
// @BasePath /
import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"vlrggapi/internal/router"

	// Explicitly import all handlers for swag to find them
	_ "vlrggapi/internal/scrapers"
	_ "vlrggapi/docs"

	"github.com/gofiber/swagger"
	"go.uber.org/zap"
	"sync"
)

func main() {
	// Initialize zap logger
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()

	app := fiber.New(fiber.Config{
		AppName:      "vlrggapi",
		ServerHeader: "vlrggapi",
	})

	// Enable CORS for all origins and methods
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "*",
	}))

	// Fiber logger middleware (console)
	app.Use(logger.New())

	// Global rate limiting (600 requests/minute per IP)
	app.Use(limiter.New(limiter.Config{
		Max:        600,
		Expiration: 60 * 1000 * 1000 * 1000, // 1 minute in nanoseconds
	}))

	// Simple in-memory cache for GET requests (per endpoint+query)
	type cacheEntry struct {
		data      []byte
		timestamp time.Time
	}
	var (
		cache      = make(map[string]cacheEntry)
		cacheMutex sync.RWMutex
		cacheTTL   = 30 * time.Second // cache duration
	)
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodGet {
			return c.Next()
		}
		key := c.OriginalURL()
		cacheMutex.RLock()
		entry, found := cache[key]
		cacheMutex.RUnlock()
		if found && time.Since(entry.timestamp) < cacheTTL {
			c.Response().Header.Set("X-Cache", "HIT")
			return c.Send(entry.data)
		}
		// Capture response
		err := c.Next()
		if err != nil {
			loggerZap.Error("handler error", zap.String("url", key), zap.Error(err))
			return err
		}
		if c.Response().StatusCode() == fiber.StatusOK {
			cacheMutex.Lock()
			cache[key] = cacheEntry{
				data:      c.Response().Body(),
				timestamp: time.Now(),
			}
			cacheMutex.Unlock()
			c.Response().Header.Set("X-Cache", "MISS")
		}
		return nil
	})

	// Register VLR router
	router.RegisterVlrRoutes(app)

	// Root redirect to docs
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs", fiber.StatusFound)
	})

	// Swagger UI endpoint
	app.Get("/swagger/*", swagger.HandlerDefault) // default: /swagger/index.html

	// Simple /docs endpoint (legacy/redirect)
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusFound)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	loggerZap.Info("Starting server", zap.String("port", port))
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
