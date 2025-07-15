package scrapers

import (
	"net/http"
	"strings"

	"vlrggapi/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
)

/*
@Summary      Get latest Valorant news
@Description  Returns a list of recent Valorant news articles
@Tags         news
@Produce      json
@Success      200  {object}  map[string]interface{}
@Failure      500  {object}  map[string]string
@Router       /vlr/news [get]
*/
func VlrNews(c *fiber.Ctx) error {
	url := "https://www.vlr.gg/news"
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range utils.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch news"})
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse HTML"})
	}

	var result []map[string]string
	doc.Find("a.wf-module-item").Each(func(i int, s *goquery.Selection) {
		dateAuthor := s.Find("div.ge-text-light").Text()
		parts := strings.Split(dateAuthor, "by")
		date := ""
		author := ""
		if len(parts) == 2 {
			date = strings.TrimSpace(strings.Split(parts[0], "•")[1])
			author = strings.TrimSpace(parts[1])
		}
		// Title: get the first direct text node of the module, which is the news title
		title := ""
		s.Find("div").EachWithBreak(func(i int, div *goquery.Selection) bool {
			// The first div with non-empty text and not ge-text-light is the title
			class, _ := div.Attr("class")
			text := strings.TrimSpace(div.Text())
			if class != "ge-text-light" && text != "" {
				title = strings.Split(text, "\n")[0]
				title = strings.ReplaceAll(title, "\t", "")
				title = strings.TrimSpace(title)
				return false // break
			}
			return true
		})

		// Description: second div inside the module
		desc := s.Find("div").Find("div:nth-child(2)").Text()
		desc = strings.ReplaceAll(desc, "\n\n\t\t\t\t\t", "")
		desc = strings.ReplaceAll(desc, "\t", "")
		desc = strings.ReplaceAll(desc, "\n", "")
		desc = strings.TrimSpace(desc)

		urlPath, _ := s.Attr("href")
		result = append(result, map[string]string{
			"title":       title,
			"description": desc,
			"date":        date,
			"author":      author,
			"url_path":    "https://vlr.gg" + urlPath,
		})
	})

	return c.JSON(fiber.Map{"data": fiber.Map{"status": resp.StatusCode, "segments": result}})
}
