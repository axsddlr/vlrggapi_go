package scrapers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"vlrggapi/internal/utils"
)

import "math"

func pow(a float64, b int) float64 {
	return math.Pow(a, float64(b))
}

//
// VlrLiveScore godoc
// @Summary      Get live Valorant match scores
// @Description  Returns live match scores from VLR.GG
// @Tags         matches
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /vlr/live [get]
//
func VlrLiveScore(c *fiber.Ctx) error {
	url := "https://www.vlr.gg"
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range utils.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch live matches"})
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse HTML"})
	}

	var result []map[string]interface{}
	doc.Find(".js-home-matches-upcoming a.wf-module-item").Each(func(_ int, s *goquery.Selection) {
		isLive := s.Find(".h-match-eta.mod-live")
		if isLive.Length() > 0 {
			teams := []string{}
			flags := []string{}
			scores := []string{}
			roundTexts := []map[string]string{}
			s.Find(".h-match-team").Each(func(_ int, team *goquery.Selection) {
				teams = append(teams, strings.TrimSpace(team.Find(".h-match-team-name").Text()))
				flagClass, _ := team.Find(".flag").Attr("class")
				flagClass = strings.ReplaceAll(flagClass, " mod-", "")
				flagClass = strings.ReplaceAll(flagClass, "16", "_")
				flags = append(flags, flagClass)
				scores = append(scores, strings.TrimSpace(team.Find(".h-match-team-score").Text()))
				roundInfoCT := team.Find(".h-match-team-rounds .mod-ct")
				roundInfoT := team.Find(".h-match-team-rounds .mod-t")
				roundTextCT := "N/A"
				roundTextT := "N/A"
				if roundInfoCT.Length() > 0 {
					roundTextCT = strings.TrimSpace(roundInfoCT.First().Text())
				}
				if roundInfoT.Length() > 0 {
					roundTextT = strings.TrimSpace(roundInfoT.First().Text())
				}
				roundTexts = append(roundTexts, map[string]string{"ct": roundTextCT, "t": roundTextT})
			})

			eta := "LIVE"
			matchEvent := strings.TrimSpace(s.Find(".h-match-preview-event").Text())
			matchSeries := strings.TrimSpace(s.Find(".h-match-preview-series").Text())
			timestamp := ""
			if ts, exists := s.Find(".moment-tz-convert").Attr("data-utc-ts"); exists {
				sec, _ := strconv.ParseInt(ts, 10, 64)
				timestamp = time.Unix(sec, 0).UTC().Format("2006-01-02 15:04:05")
			}
			urlPath, _ := s.Attr("href")
			urlPath = "https://www.vlr.gg" + urlPath

			// Fetch match page for team logos and map info
			teamLogos := []string{"", ""}
			currentMap := "Unknown"
			mapNumber := "Unknown"
			matchPageReq, _ := http.NewRequest("GET", urlPath, nil)
			for k, v := range utils.Headers {
				matchPageReq.Header.Set(k, v)
			}
			matchPageResp, err := http.DefaultClient.Do(matchPageReq)
			if err == nil {
				defer matchPageResp.Body.Close()
				matchDoc, err := goquery.NewDocumentFromReader(matchPageResp.Body)
				if err == nil {
					matchDoc.Find(".match-header-vs img").Each(func(i int, img *goquery.Selection) {
						if i < 2 {
							src, _ := img.Attr("src")
							teamLogos[i] = "https:" + src
						}
					})
					activeMap := matchDoc.Find(".vm-stats-gamesnav-item.js-map-switch.mod-active.mod-live")
					if activeMap.Length() > 0 {
						mapDiv := activeMap.Find("div")
						if mapDiv.Length() > 0 {
							mapText := strings.TrimSpace(mapDiv.Text())
							mapText = strings.ReplaceAll(mapText, "\n", "")
							mapText = strings.ReplaceAll(mapText, "\t", "")
							currentMap = mapText
							re := regexp.MustCompile(`^\d+`)
							mapNumberMatch := re.FindString(mapText)
							if mapNumberMatch != "" {
								mapNumber = mapNumberMatch
								currentMap = strings.TrimSpace(strings.TrimPrefix(mapText, mapNumberMatch))
							}
						}
					}
				}
			}

			team1RoundCT := "N/A"
			team1RoundT := "N/A"
			team2RoundCT := "N/A"
			team2RoundT := "N/A"
			if len(roundTexts) > 0 {
				team1RoundCT = roundTexts[0]["ct"]
				team1RoundT = roundTexts[0]["t"]
			}
			if len(roundTexts) > 1 {
				team2RoundCT = roundTexts[1]["ct"]
				team2RoundT = roundTexts[1]["t"]
			}

			result = append(result, map[string]interface{}{
				"team1":           teams[0],
				"team2":           teams[1],
				"flag1":           flags[0],
				"flag2":           flags[1],
				"team1_logo":      teamLogos[0],
				"team2_logo":      teamLogos[1],
				"score1":          scores[0],
				"score2":          scores[1],
				"team1_round_ct":  team1RoundCT,
				"team1_round_t":   team1RoundT,
				"team2_round_ct":  team2RoundCT,
				"team2_round_t":   team2RoundT,
				"map_number":      mapNumber,
				"current_map":     currentMap,
				"time_until_match": eta,
				"match_event":     matchEvent,
				"match_series":    matchSeries,
				"unix_timestamp":  timestamp,
				"match_page":      urlPath,
			})
		}
	})

	// If no live matches, add a message
	if len(result) == 0 {
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"status": 200,
				"segments": []interface{}{},
				"message": "No live matches at this time.",
			},
		})
	}

	return c.JSON(fiber.Map{"data": fiber.Map{"status": 200, "segments": result}})
}

//
// VlrMatchResults godoc
// @Summary      Get recent Valorant match results
// @Description  Returns recent match results with detailed info
// @Tags         matches
// @Produce      json
// @Param        num_pages     query     int     false  "Number of pages to fetch"  default(1)
// @Param        from_page     query     int     false  "Start page"
// @Param        to_page       query     int     false  "End page"
// @Param        max_retries   query     int     false  "Retry attempts per page"    default(3)
// @Param        request_delay query     number  false  "Delay between requests (seconds)" default(1.0)
// @Param        timeout       query     int     false  "HTTP timeout (seconds)"     default(30)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /vlr/match [get]
//
func VlrMatchResults(c *fiber.Ctx) error {
	numPages, _ := strconv.Atoi(c.Query("num_pages", "1"))
	fromPageStr := c.Query("from_page")
	toPageStr := c.Query("to_page")
	maxRetries, _ := strconv.Atoi(c.Query("max_retries", "3"))
	requestDelay, _ := strconv.ParseFloat(c.Query("request_delay", "1.0"), 64)
	timeout, _ := strconv.Atoi(c.Query("timeout", "30"))

	var fromPage, toPage int
	var err error
	if fromPageStr != "" {
		fromPage, err = strconv.Atoi(fromPageStr)
		if err != nil || fromPage < 1 {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid from_page"})
		}
	}
	if toPageStr != "" {
		toPage, err = strconv.Atoi(toPageStr)
		if err != nil || toPage < 1 {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid to_page"})
		}
	}

	// Determine page range
	startPage, endPage, totalPages := 1, numPages, numPages
	if fromPageStr != "" && toPageStr != "" {
		startPage = fromPage
		endPage = toPage
		totalPages = endPage - startPage + 1
	} else if fromPageStr != "" {
		startPage = fromPage
		endPage = fromPage + numPages - 1
		totalPages = numPages
	} else if toPageStr != "" {
		startPage = toPage - numPages + 1
		if startPage < 1 {
			startPage = 1
		}
		endPage = toPage
		totalPages = endPage - startPage + 1
	}

	type MatchResult struct {
		Team1          string `json:"team1"`
		Team2          string `json:"team2"`
		Score1         string `json:"score1"`
		Score2         string `json:"score2"`
		Flag1          string `json:"flag1"`
		Flag2          string `json:"flag2"`
		TimeCompleted  string `json:"time_completed"`
		RoundInfo      string `json:"round_info"`
		TournamentName string `json:"tournament_name"`
		MatchPage      string `json:"match_page"`
		TournamentIcon string `json:"tournament_icon"`
		PageNumber     int    `json:"page_number"`
	}

	var result []MatchResult
	var failedPages []int

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	for page := startPage; page <= endPage; page++ {
		pageSuccess := false
		retryCount := 0

		for !pageSuccess && retryCount < maxRetries {
			var url string
			if page == 1 {
				url = "https://www.vlr.gg/matches/results"
			} else {
				url = fmt.Sprintf("https://www.vlr.gg/matches/results/?page=%d", page)
			}

			req, _ := http.NewRequest("GET", url, nil)
			for k, v := range utils.Headers {
				req.Header.Set(k, v)
			}
			resp, err := client.Do(req)
			if err != nil {
				retryCount++
				if retryCount < maxRetries {
					sleepDuration := time.Duration(float64(requestDelay) * float64(time.Second) * pow(2, retryCount))
					time.Sleep(sleepDuration)
				}
				continue
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				retryCount++
				if retryCount < maxRetries {
					sleepDuration := time.Duration(float64(requestDelay) * float64(time.Second) * pow(2, retryCount))
					time.Sleep(sleepDuration)
				}
				continue
			}

			items := doc.Find("a.wf-module-item")
			if items.Length() == 0 {
				pageSuccess = true
				break
			}

			items.Each(func(_ int, s *goquery.Selection) {
				urlPath, _ := s.Attr("href")

				// Parse team names and scores from the new HTML structure
				vs := s.Find(".match-item-vs")
				team1 := ""
				team2 := ""
				score1 := ""
				score2 := ""
				flag1 := ""
				flag2 := ""

				teams := vs.Find(".match-item-vs-team")
				if teams.Length() >= 2 {
					team1Div := teams.Eq(0)
					team2Div := teams.Eq(1)

					team1 = strings.TrimSpace(team1Div.Find(".match-item-vs-team-name .text-of").Text())
					team2 = strings.TrimSpace(team2Div.Find(".match-item-vs-team-name .text-of").Text())
					score1 = strings.TrimSpace(team1Div.Find(".match-item-vs-team-score").Text())
					score2 = strings.TrimSpace(team2Div.Find(".match-item-vs-team-score").Text())

					flag1Sel := team1Div.Find(".match-item-vs-team-name .flag")
					flag2Sel := team2Div.Find(".match-item-vs-team-name .flag")
					if flag1Sel.Length() > 0 {
						class, _ := flag1Sel.Attr("class")
						class = strings.ReplaceAll(class, " mod-", "_")
						flag1 = class
					}
					if flag2Sel.Length() > 0 {
						class, _ := flag2Sel.Attr("class")
						class = strings.ReplaceAll(class, " mod-", "_")
						flag2 = class
					}
				}

				// Fallback for time completed and event info
				divs := s.Find("div")
				clean := func(str string) string {
					return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(str, "\t", ""), "\n", ""))
				}
				timeCompleted := clean(divs.Eq(0).Text())

				roundInfo := ""
				tournamentName := ""
				tournamentIcon := ""
				divs.Each(func(i int, d *goquery.Selection) {
					class, _ := d.Attr("class")
					if strings.Contains(class, "match-item-event-series") {
						roundInfo = clean(d.Text())
					}
					if strings.Contains(class, "match-item-event") {
						tournamentName = clean(d.Text())
						img := d.Find("img")
						if img.Length() > 0 {
							tournamentIcon, _ = img.Attr("src")
							if !strings.HasPrefix(tournamentIcon, "http") {
								tournamentIcon = "https:" + tournamentIcon
							}
						}
					}
				})

				result = append(result, MatchResult{
					Team1:          team1,
					Team2:          team2,
					Score1:         score1,
					Score2:         score2,
					Flag1:          flag1,
					Flag2:          flag2,
					TimeCompleted:  timeCompleted,
					RoundInfo:      roundInfo,
					TournamentName: tournamentName,
					MatchPage:      urlPath,
					TournamentIcon: tournamentIcon,
					PageNumber:     page,
				})
			})

			pageSuccess = true
			if page < endPage {
				time.Sleep(time.Duration(requestDelay * float64(time.Second)))
			}
		}

		if !pageSuccess {
			failedPages = append(failedPages, page)
		}
	}

	segments := fiber.Map{
		"status": 200,
		"segments": result,
		"meta": fiber.Map{
			"page_range":            fmt.Sprintf("%d-%d", startPage, endPage),
			"total_pages_requested": totalPages,
			"successful_pages":      totalPages - len(failedPages),
			"failed_pages":          failedPages,
			"total_matches":         len(result),
		},
	}
	data := fiber.Map{"data": segments}
	if len(result) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("No data retrieved. Failed pages: %v", failedPages)})
	}
	return c.JSON(data)
}
