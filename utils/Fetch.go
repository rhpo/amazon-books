package utils

import (
	"fmt"

	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
)

func Fetch(url string) (string, error) {
	client, err := cloudscraper.Init(false, false)
	if err != nil {
		return "", fmt.Errorf("failed to create CloudScraper client: %w", err)
	}

	res, err := client.Get(url, make(map[string]string), "")
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}

	return res.Body, nil
}

func Report(message string) error {
	println(message)
	return fmt.Errorf(message)
}
