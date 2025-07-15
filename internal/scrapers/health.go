package scrapers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

//
// Health godoc
// @Summary      Health check
// @Description  Returns health status of the API and upstream sources
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /vlr/health [get]
//
func Health(c *fiber.Ctx) error {
	sites := []string{"https://vlrggapi.vercel.app", "https://vlr.gg"}
	results := make(map[string]map[string]interface{})
	for _, site := range sites {
		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(site)
		status := "Unhealthy"
		statusCode := 0
		if err == nil {
			if resp.StatusCode == 200 {
				status = "Healthy"
			}
			statusCode = resp.StatusCode
			resp.Body.Close()
		}
		results[site] = map[string]interface{}{
			"status":      status,
			"status_code": statusCode,
		}
	}
	return c.JSON(results)
}
