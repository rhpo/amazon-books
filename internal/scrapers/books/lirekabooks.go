package books

import (
	"amazon/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Internal response structs
type algoliaResponse struct {
	Results []struct {
		Hits []map[string]any `json:"hits"`
	} `json:"results"`
}

// LirekaSearchBooks fetches and parses Lireka’s Algolia API results.
func LirekaSearchBooks(query string) ([]models.Book, error) {
	urlStr := "https://mwx92vzv2w-dsn.algolia.net/1/indexes/*/queries"

	params := url.Values{}
	params.Set("x-algolia-agent", "Algolia for JavaScript (3.35.1)")
	params.Set("x-algolia-application-id", "MWX92VZV2W")
	params.Set("x-algolia-api-key", "b99c00173786225dd85f6ede7ccd003e")

	requestBody := map[string]any{
		"requests": []map[string]string{
			{
				"indexName": "books",
				"params": fmt.Sprintf(
					"query=%s&hitsPerPage=40&page=0"+
						"&filters=channels:11 AND NOT suppliedByLireka:true"+
						"&clickAnalytics=true",
					url.QueryEscape(query),
				),
			},
		},
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", urlStr+"?"+params.Encode(), bytes.NewBuffer(body))
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.lireka.com/")
	req.Header.Set("Origin", "https://www.lireka.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var algRes algoliaResponse
	if err := json.Unmarshal(data, &algRes); err != nil {
		return nil, err
	}

	if len(algRes.Results) == 0 {
		return nil, nil
	}

	var books []models.Book
	hits := algRes.Results[0].Hits

	for _, h := range hits {
		book := models.Book{
			ID:          getString(h["objectID"]),
			Title:       getString(h["title"]),
			Description: getString(h["description"]),
			Publisher:   getString(h["imprints"]),
			PubDate:     getString(h["publicationDateStr"]),
			IsGBook:     false,
		}

		// Image handling (Lireka encodes paths)
		if imgs, ok := h["images"].([]interface{}); ok && len(imgs) > 0 {
			code := getString(imgs[0])
			book.Cover = fmt.Sprintf("https://media.lireka.com/%s?resize=fit&h=900&w=600&gq=1&v=1", code)
		}

		// Pages
		if pages, ok := h["numberOfPages"].(float64); ok {
			book.Pages = int(pages)
		}

		// Price — take the first available listing
		if listings, ok := h["listings_lireka"].(map[string]interface{}); ok {
			for _, v := range listings {
				if lmap, ok := v.(map[string]any); ok {
					if priceVal, ok := lmap["price"].(float64); ok {
						book.Price = float32(priceVal)
					}
					break
				}
			}
		}

		// Rating
		if rating, ok := h["rating"].(float64); ok {
			book.Rating = float32(rating)
		} else if ratingInt, ok := h["ratingInt"].(float64); ok {
			book.Rating = float32(ratingInt)
		}

		// Authors
		if authors, ok := h["authors"].([]any); ok {
			for _, a := range authors {
				if amap, ok := a.(map[string]any); ok {
					name := getString(amap["fullName"])
					name = strings.TrimSpace(name)
					if name == "" {
						name = getString(amap["name"])
						name = strings.TrimSpace(name)

					}
					if name != "" {
						book.Authors = append(book.Authors, models.AuthorType{Name: name})
					}
				}
			}
		}

		// Dimensions
		book.Dimension = models.Dimension{
			Width:  getFloat(h["width"]) / 100,
			Height: getFloat(h["height"]) / 100,
			Depth:  getFloat(h["weight"]) / 100,
		}

		book.Language = getString(h["language"])

		books = append(books, book)
	}

	return books, nil
}

func getString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func getFloat(v any) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return float64(f)
	}
	return 0
}
