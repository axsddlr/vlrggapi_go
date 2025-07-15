package scrapers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"vlrggapi/internal/utils"
)

//
// VlrStats godoc
// @Summary      Get Valorant player statistics
// @Description  Returns player statistics, filterable by region and timespan
// @Tags         stats
// @Produce      json
// @Param        region    query     string  true   "Region key (e.g. na, eu, ap, la, oce, kr, mn, gc, br, cn, jp, col)"
// @Param        timespan  query     string  false  "Timespan (e.g. all, 30 for 30 days)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /vlr/stats [get]
//
func VlrStats(c *fiber.Ctx) error {
	region := c.Query("region")
	timespan := c.Query("timespan")

	baseURL := fmt.Sprintf("https://www.vlr.gg/stats/?event_group_id=all&event_id=all&region=%s&country=all&min_rounds=200&min_rating=1550&agent=all&map_id=all", region)
	var url string
	if strings.ToLower(timespan) == "all" {
		url = baseURL + "&timespan=all"
	} else {
		url = baseURL + "&timespan=" + timespan + "d"
	}

	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range utils.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch stats"})
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse HTML"})
	}

	var result []map[string]interface{}
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		player := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(s.Text(), "\t", ""), "\n", " "))
		playerName := ""
		org := "N/A"
		if len(player) > 0 {
			playerName = player[0]
		}
		if len(player) > 1 {
			org = player[1]
		}

		var agents []string
		s.Find("td.mod-agents img").Each(func(_ int, img *goquery.Selection) {
			src, _ := img.Attr("src")
			parts := strings.Split(src, "/")
			if len(parts) > 0 {
				agent := strings.TrimSuffix(parts[len(parts)-1], ".png")
				agents = append(agents, agent)
			}
		})

		var colorSq []string
		s.Find("td.mod-color-sq").Each(func(_ int, stat *goquery.Selection) {
			colorSq = append(colorSq, stat.Text())
		})

		rnd := s.Find("td.mod-rnd").Text()

		// Defensive: check colorSq length
		for len(colorSq) < 11 {
			colorSq = append(colorSq, "")
		}

		result = append(result, map[string]interface{}{
			"player":                     playerName,
			"org":                        org,
			"agents":                     agents,
			"rounds_played":              rnd,
			"rating":                     colorSq[0],
			"average_combat_score":       colorSq[1],
			"kill_deaths":                colorSq[2],
			"kill_assists_survived_traded": colorSq[3],
			"average_damage_per_round":   colorSq[4],
			"kills_per_round":            colorSq[5],
			"assists_per_round":          colorSq[6],
			"first_kills_per_round":      colorSq[7],
			"first_deaths_per_round":     colorSq[8],
			"headshot_percentage":        colorSq[9],
			"clutch_success_percentage":  colorSq[10],
		})
	})

	return c.JSON(fiber.Map{"data": fiber.Map{"status": resp.StatusCode, "segments": result}})
}
