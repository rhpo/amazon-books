package books

import (
	"amazon/internal/utils"
	"amazon/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// -------------------- Internal Response Struct --------------------
type algoliaResponse struct {
	Results []struct {
		Hits []map[string]any `json:"hits"`
	} `json:"results"`
}

// -------------------- Public Functions --------------------

// LirekaSearchBooks fetches books from the original Lireka Algolia API
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
					"query=%s&hitsPerPage=40&page=0&filters=channels:11 AND NOT suppliedByLireka:true&clickAnalytics=true",
					url.QueryEscape(query),
				),
			},
		},
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", urlStr+"?"+params.Encode(), bytes.NewBuffer(body))
	setBasicHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseAlgoliaResponse(data)
}

// GetBooks fetches books from the newer Algolia API using the fetch request headers
func GetBooks(query string) ([]models.Book, error) {
	body := buildGetBooksBody(query)
	baseURL := "https://mwx92vzv2w-dsn.algolia.net/1/indexes/*/queries"
	params := url.Values{}
	params.Set("x-algolia-agent", "Algolia for JavaScript (3.35.1); Browser (lite); angular (13.3.11); angular-instantsearch (3.0.0-beta.5); instantsearch.js (3.5.4); JS Helper (2.28.1)")
	params.Set("x-algolia-application-id", "MWX92VZV2W")
	params.Set("x-algolia-api-key", "b99c00173786225dd85f6ede7ccd003e")

	req, _ := http.NewRequest("POST", baseURL+"?"+params.Encode(), bytes.NewBuffer(body))
	setGetBooksHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseAlgoliaResponse(respData)
}

// FetchLBooks fetches books and returns them as BookThumbnail
// Page parameter is kept in the signature but ignored internally
func FetchLBooks(page int) (*[]models.BookThumbnail, int, error) {
	// -------------------- Cache check --------------------
	cachePattern := fmt.Sprintf("%s/%d-*.json", utils.CACHE_DIRECTORY, page)
	if utils.CacheValid(cachePattern, utils.CACHE_DURATION) {
		return loadFromCache(cachePattern)
	}

	// -------------------- Fetch from Algolia --------------------
	books, err := GetBooks("") // empty query for main listing
	if err != nil {
		utils.Report("Failed to fetch books: " + err.Error())
		return loadFromCache(cachePattern) // fallback to cache if exists
	}

	if len(books) == 0 {
		return nil, 0, utils.Report("No books returned from API")
	}

	// -------------------- Convert to Thumbnails --------------------
	thumbnails := make([]models.BookThumbnail, len(books))
	for i, b := range books {
		thumbnails[i] = models.BookThumbnail{
			ID:      b.ID,
			Link:    fmt.Sprintf("/book/%s", b.ID),
			Title:   b.Title,
			Cover:   b.Cover,
			Authors: b.Authors,
			Rating:  b.Rating,
			IsGBook: b.IsGBook,
		}
	}

	// -------------------- Save to cache --------------------
	cacheFile := fmt.Sprintf("%s/%d-%d.json", utils.CACHE_DIRECTORY, page, len(thumbnails))
	if err := saveToCache(cacheFile, thumbnails); err != nil {
		utils.Report("Failed to write cache file: " + err.Error())
	}

	return &thumbnails, len(thumbnails), nil
}

// -------------------- Helpers --------------------

// Shared parser for Algolia response
func parseAlgoliaResponse(data []byte) ([]models.Book, error) {
	var res algoliaResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if len(res.Results) == 0 {
		return nil, nil
	}

	hits := res.Results[0].Hits
	var books []models.Book
	for _, h := range hits {
		books = append(books, parseBookHit(h))
	}
	return books, nil
}

// parseBookHit converts a single hit into a Book
func parseBookHit(h map[string]any) models.Book {
	book := models.Book{
		ID:          getString(h["objectID"]),
		Title:       getString(h["title"]),
		Description: getString(h["description"]),
		Publisher:   getString(h["imprints"]),
		PubDate:     getString(h["publicationDateStr"]),
		IsGBook:     false,
		Dimension: models.Dimension{
			Width:  getFloat(h["width"]) / 100,
			Height: getFloat(h["height"]) / 100,
			Depth:  getFloat(h["weight"]) / 100,
		},
	}

	// Language
	if langs, ok := h["languages"].([]any); ok && len(langs) > 0 {
		book.Language = getString(langs[0])
	}

	// Pages
	if pages, ok := h["numberOfPages"].(float64); ok {
		book.Pages = int(pages)
	} else if pagesStr := getString(h["numberOfPagesStr"]); pagesStr != "" {
		fmt.Sscanf(pagesStr, "%d", &book.Pages)
	}

	// Cover image
	if imgs, ok := h["images"].([]any); ok && len(imgs) > 0 {
		book.Cover = fmt.Sprintf("https://media.lireka.com/%s?resize=fit&h=900&w=600&gq=1&v=1", getString(imgs[0]))
	}

	// Price
	if listings, ok := h["listings"].(map[string]any); ok {
		for _, v := range listings {
			if lmap, ok := v.(map[string]any); ok {
				if price, ok := lmap["price"].(float64); ok {
					book.Price = float32(price)
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
				name := strings.TrimSpace(getString(amap["fullName"]))
				if name == "" {
					name = strings.TrimSpace(getString(amap["lastName"]))
				}
				if name != "" {
					book.Authors = append(book.Authors, models.AuthorType{Name: name})
				}
			}
		}
	}

	return book
}

// -------------------- Request Builders --------------------

// buildGetBooksBody constructs the JSON body for the new fetch request
func buildGetBooksBody(query string) []byte {
	reqMap := map[string]any{
		"requests": []map[string]string{
			{
				"indexName": "books",
				"params": fmt.Sprintf(
					"query=%s&hitsPerPage=40&maxValuesPerFacet=50&page=0&highlightPreTag=__ais-highlight__&highlightPostTag=__%%2Fais-highlight__"+
						"&optionalFilters=[\"has_image:-false\",\"availability:-UNAVAILABLE\",\"listings_lireka.1.availability:-UNAVAILABLE\",\"listings.1.availability:-UNAVAILABLE\",\"listings_lireka.1.availability:IN_TWO_WEEKS\",\"listings_lireka.2.availability:-UNAVAILABLE\",\"availability:-POD\",\"bookForm:-AUDIO_DISC\",\"listings_lireka.1.availability:POD\",\"amazonSalesRankingLevel:A\",\"amazonSalesRankingLevel:B\",\"languages:fre\"]"+
						"&clickAnalytics=true&userToken=201475150"+
						"&filters=bookSelections:111 AND channels:11 AND NOT listings_lireka.11.availability:UNAVAILABLE AND NOT suppliedByLireka:true"+
						"&facets=[\"listings_lireka.11.availability\",\"categories_level_1.name.fr\",\"categories_level_2.name.fr\",\"ratingInt\",\"listings_lireka.11.price\",\"bookForm\",\"languages\"]"+
						"&tagFilters=",
					url.QueryEscape(query),
				),
			},
		},
	}

	body, _ := json.Marshal(reqMap)
	return body
}

// -------------------- Headers --------------------
func setBasicHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.lireka.com/")
	req.Header.Set("Origin", "https://www.lireka.com")
}

func setGetBooksHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.6")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-CH-UA", `"Not:A-Brand";v="99", "Brave";v="145", "Chromium";v="145"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Referer", "https://www.lireka.com/")
}

// -------------------- Utilities --------------------
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
		return f
	}
	return 0
}
