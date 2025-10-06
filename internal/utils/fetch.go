package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const scrapingBotAPI = "http://api.scraping-bot.io/scrape/raw-html"

func Fetch(targetURL string) (string, int, error, bool) {
	username := os.Getenv("SCRAPING_BOT_USER")
	apiKey := os.Getenv("SCRAPING_BOT_KEY")
	if username == "" || apiKey == "" {
		return "", 500, fmt.Errorf("SCRAPING_BOT_USER or SCRAPING_BOT_KEY env vars not set"), true
	}

	// Prepare JSON body
	payload := map[string]any{
		"url": targetURL,
		"options": map[string]any{
			"useChrome":              false,
			"premiumProxy":           false,
			"proxyCountry":           nil,
			"waitForNetworkRequests": false,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", 500, fmt.Errorf("failed to marshal payload: %w", err), true
	}

	// Prepare request
	req, err := http.NewRequest("POST", scrapingBotAPI, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", 500, fmt.Errorf("failed to create request: %w", err), true
	}
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + apiKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 500, fmt.Errorf("failed to fetch via scraping-bot.io: %w", err), true
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 500, fmt.Errorf("failed to read response body: %w", err), true
	}

	println(string(respBody))

	return string(respBody), resp.StatusCode, nil, false
}
