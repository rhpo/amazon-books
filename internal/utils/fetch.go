package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
)

const maxTry = 5
const failedMsg = "Request was throttled. Please wait a moment and refresh the page"

func Fetch(url string) (string, error) {
	return fetchWithRetries(url, 0)
}

func fetchWithRetries(url string, attempt int) (string, error) {
	if attempt > maxTry {
		return "", Report("Max retries exceeded while fetching: " + url)
	}

	client, err := cloudscraper.Init(false, false)
	if err != nil {
		return "", Report("failed to create CloudScraper client: " + err.Error())
	}

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.5",
	}

	res, err := client.Get(url, headers, "")
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}

	if len(res.Body) < 70 && strings.Contains(res.Body, "wait a moment and refresh the page") {
		// wait 2 to 3 seconds
		rand.Seed(time.Now().UnixNano())
		delay := time.Duration(1+rand.Intn(2)) * time.Second
		fmt.Printf("Request throttled. Waiting %v before retrying...\n", delay)
		time.Sleep(delay)

		return fetchWithRetries(url, attempt+1)
	}

	return res.Body, nil
}
