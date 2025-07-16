package scrapers

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"vlrggapi/internal/utils"
)

//
// VlrEvents godoc
// @Summary      Get Valorant events
// @Description  Returns a list of upcoming or completed Valorant events
// @Tags         events
// @Produce      json
// @Param        upcoming   query     bool  false  "If true, return only upcoming events"
// @Param        completed  query     bool  false  "If true, return only completed events"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /vlr/events [get]
//
func VlrEvents(c *fiber.Ctx) error {
	url := "https://www.vlr.gg/events"
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range utils.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse HTML"})
	}

	// Query params
	showUpcoming := c.Query("upcoming") != "false"
	showCompleted := c.Query("completed") != "false"

	// If both are explicitly false, show both (default)
	if !showUpcoming && !showCompleted {
		showUpcoming = true
		showCompleted = true
	}

	events := []map[string]string{}

	// Helper to parse event cards
	parseEvents := func(sel *goquery.Selection) {
		sel.Find("a.event-item").Each(func(_ int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Find(".event-item-title").Text())
			status := strings.TrimSpace(s.Find(".event-item-desc-item-status").Text())
			prize := strings.TrimSpace(s.Find(".event-item-desc-item.mod-prize").Clone().Children().Remove().End().Text())
			dates := strings.TrimSpace(s.Find(".event-item-desc-item.mod-dates").Clone().Children().Remove().End().Text())
			region := ""
			flag := s.Find(".event-item-desc-item.mod-location .flag")
			if flag.Length() > 0 {
				class, _ := flag.Attr("class")
				region = strings.TrimSpace(strings.ReplaceAll(class, "flag mod-", ""))
			}
			thumb := ""
			img := s.Find(".event-item-thumb img")
			if img.Length() > 0 {
				src, _ := img.Attr("src")
				if strings.HasPrefix(src, "//") {
					thumb = "https:" + src
				} else if strings.HasPrefix(src, "/") {
					thumb = "https://www.vlr.gg" + src
				} else {
					thumb = src
				}
			}
			urlPath, _ := s.Attr("href")
			events = append(events, map[string]string{
				"title":    title,
				"status":   status,
				"prize":    prize,
				"dates":    dates,
				"region":   region,
				"thumb":    thumb,
				"url_path": "https://www.vlr.gg" + urlPath,
			})
		})
	}

	// Upcoming events
	if showUpcoming {
		doc.Find("div.wf-label.mod-large.mod-upcoming").Each(func(_ int, s *goquery.Selection) {
			// The next sibling is the container for upcoming events
			upcomingCol := s.Parent().Find("a.event-item")
			if upcomingCol.Length() > 0 {
				parseEvents(s.Parent())
			}
		})
	}

	// Completed events
	if showCompleted {
		doc.Find("div.wf-label.mod-large.mod-completed").Each(func(_ int, s *goquery.Selection) {
			completedCol := s.Parent().Find("a.event-item")
			if completedCol.Length() > 0 {
				parseEvents(s.Parent())
			}
		})
	}

	return c.JSON(fiber.Map{"data": fiber.Map{"status": resp.StatusCode, "segments": events}})
}
