package scrapers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"vlrggapi/internal/utils"
)

//
// VlrRankings godoc
// @Summary      Get Valorant team rankings
// @Description  Returns team rankings for a given region
// @Tags         rankings
// @Produce      json
// @Param        region  query     string  true   "Region key (e.g. na, eu, ap, la, oce, kr, mn, gc, br, cn, jp, col)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /vlr/rankings [get]
//
func VlrRankings(c *fiber.Ctx) error {
	regionKey := c.Query("region")
	regionMap := utils.Region
	regionVal, ok := regionMap[regionKey]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid region"})
	}
	url := "https://www.vlr.gg/rankings/" + regionVal

	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range utils.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch rankings"})
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse HTML"})
	}

	var result []map[string]interface{}
	doc.Find("div.rank-item").Each(func(i int, s *goquery.Selection) {
		rank := strings.TrimSpace(s.Find("div.rank-item-rank-num").Text())
		team := strings.Split(s.Find("div.ge-text").Text(), "#")[0]
		logo := s.Find("a.rank-item-team").Find("img").AttrOr("src", "")
		re := regexp.MustCompile(`/img/vlr/tmp/vlr.png`)
		logo = re.ReplaceAllString(logo, "")
		country := s.Find("div.rank-item-team-country").Text()
		lastPlayed := strings.Split(strings.ReplaceAll(strings.ReplaceAll(s.Find("a.rank-item-last").Text(), "\n", ""), "\t", ""), "v")[0]
		lastPlayedTeamRaw := strings.ReplaceAll(strings.ReplaceAll(s.Find("a.rank-item-last").Text(), "\t", ""), "\n", "")
		lastPlayedTeamParts := strings.SplitN(lastPlayedTeamRaw, "o", 2)
		lastPlayedTeamStr := ""
		if len(lastPlayedTeamParts) == 2 {
			lastPlayedTeamStr = strings.ReplaceAll(lastPlayedTeamParts[1], ".", ". ")
		}
		lastPlayedTeamLogo := s.Find("a.rank-item-last").Find("img").AttrOr("src", "")
		record := strings.ReplaceAll(strings.ReplaceAll(s.Find("div.rank-item-record").Text(), "\t", ""), "\n", "")
		earnings := strings.ReplaceAll(strings.ReplaceAll(s.Find("div.rank-item-earnings").Text(), "\t", ""), "\n", "")

		result = append(result, map[string]interface{}{
			"rank":                rank,
			"team":                strings.TrimSpace(team),
			"country":             country,
			"last_played":         strings.TrimSpace(lastPlayed),
			"last_played_team":    strings.TrimSpace(lastPlayedTeamStr),
			"last_played_team_logo": lastPlayedTeamLogo,
			"record":              record,
			"earnings":            earnings,
			"logo":                logo,
		})
	})

	return c.JSON(fiber.Map{"status": resp.StatusCode, "data": result})
}
