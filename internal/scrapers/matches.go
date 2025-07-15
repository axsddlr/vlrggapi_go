package scrapers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"vlrggapi/internal/utils"
)

import "math"

func pow(a float64, b int) float64 {
	return math.Pow(a, float64(b))
}

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
		Team1            string `json:"team1"`
		Team2            string `json:"team2"`
		Score1           string `json:"score1"`
		Score2           string `json:"score2"`
		Flag1            string `json:"flag1"`
		Flag2            string `json:"flag2"`
		TimeCompleted    string `json:"time_completed"`
		RoundInfo        string `json:"round_info"`
		TournamentName   string `json:"tournament_name"`
		MatchPage        string `json:"match_page"`
		TournamentIcon   string `json:"tournament_icon"`
		PageNumber       int    `json:"page_number"`
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
					// Use math.Pow for exponential backoff, since bit shifting float64 is invalid
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
					// Use math.Pow for exponential backoff, since bit shifting float64 is invalid
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
				eta := s.Find("div.ml-eta").Text() + " ago"
				rounds := s.Find("div.match-item-event-series").Text()
				rounds = strings.ReplaceAll(rounds, "\u2013", "-")
				rounds = strings.ReplaceAll(rounds, "\n", "")
				rounds = strings.ReplaceAll(rounds, "\t", "")

				tourney := s.Find("div.match-item-event").Text()
				tourney = strings.ReplaceAll(tourney, "\t", " ")
				tourney = strings.TrimSpace(tourney)
				tourneyLines := strings.Split(tourney, "\n")
				tourneyName := ""
				if len(tourneyLines) > 1 {
					tourneyName = strings.TrimSpace(tourneyLines[1])
				}

				tourneyIconURL := ""
				img := s.Find("img")
				if img.Length() > 0 {
					tourneyIconURL, _ = img.Attr("src")
					if !strings.HasPrefix(tourneyIconURL, "http") {
						tourneyIconURL = "https:" + tourneyIconURL
					}
				}

				teamArray := ""
				vs := s.Find("div.match-item-vs").Find("div:nth-child(2)")
				if vs.Length() > 0 {
					teamArray = vs.Text()
				} else {
					teamArray = "TBD"
				}
				teamArray = strings.ReplaceAll(teamArray, "\t", " ")
				teamArray = strings.ReplaceAll(teamArray, "\n", " ")
				teamArray = strings.TrimSpace(teamArray)
				teamSplit := strings.Split(teamArray, "                                  ")
				team1, score1, team2, score2 := "", "", "", ""
				if len(teamSplit) >= 5 {
					team1 = teamSplit[0]
					score1 = strings.TrimSpace(teamSplit[1])
					team2 = teamSplit[4]
					score2 = strings.TrimSpace(teamSplit[len(teamSplit)-1])
				}

				flagList := []string{}
				s.Find(".flag").Each(func(_ int, flagParent *goquery.Selection) {
					class, _ := flagParent.Attr("class")
					class = strings.ReplaceAll(class, " mod-", "_")
					flagList = append(flagList, class)
				})
				flag1, flag2 := "", ""
				if len(flagList) > 0 {
					flag1 = flagList[0]
				}
				if len(flagList) > 1 {
					flag2 = flagList[1]
				}

				result = append(result, MatchResult{
					Team1:          team1,
					Team2:          team2,
					Score1:         score1,
					Score2:         score2,
					Flag1:          flag1,
					Flag2:          flag2,
					TimeCompleted:  eta,
					RoundInfo:      rounds,
					TournamentName: tourneyName,
					MatchPage:      urlPath,
					TournamentIcon: tourneyIconURL,
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
