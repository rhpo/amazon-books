package books

import (
	"amazon/internal/utils"
	"amazon/models"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const COVER_IMG_SIZE = 1200

func parseDocument(content string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(content))
}

func extractBookData(book *goquery.Selection, isSearch bool) (models.BookThumbnail, error) {
	var bookThumbnail models.BookThumbnail

	// Skip audio books
	bookText := strings.ToLower(book.Text())
	if utils.IsAudible(bookText) {
		return bookThumbnail, fmt.Errorf("skip audio book")
	}

	// Extract ID
	bookID, exists := book.Attr("data-asin")
	if !exists {
		return bookThumbnail, utils.Report("Can't find book ID")
	}
	bookThumbnail.ID = bookID

	// Extract Link and Title
	if err := extractLinkAndTitle(&bookThumbnail, book, isSearch); err != nil {
		return bookThumbnail, err
	}

	// Extract Cover
	if err := extractCover(&bookThumbnail, book); err != nil {
		return bookThumbnail, err
	}

	// Extract Authors
	if isSearch {
		extractAuthorsFromSearch(&bookThumbnail, book)
	} else {
		extractAuthors(&bookThumbnail, book)
	}

	// Extract Rating
	extractRating(&bookThumbnail, book)

	return bookThumbnail, nil
}

func extractLinkAndTitle(bookThumbnail *models.BookThumbnail, book *goquery.Selection, isSearch bool) error {
	var linkEl *goquery.Selection

	if isSearch {
		// Try multiple selectors for search results
		linkEl = book.Find("h2 a")
		if linkEl.Length() == 0 {
			linkEl = book.Find("a[href*='/dp/']")
		}
		if linkEl.Length() == 0 {
			linkEl = book.Find("a.a-link-normal").First()
		}
	} else {
		linkEl = book.Find(".a-link-normal.aok-block")
	}

	if linkEl.Length() == 0 {
		return utils.Report("Can't find book link element")
	}

	link, exists := linkEl.Attr("href")
	if !exists {
		return utils.Report("Can't find book link href")
	}
	bookThumbnail.Link = link

	// Extract title
	var title string
	if isSearch {
		// Try to get title from h2 span first
		titleEl := book.Find("h2 span")
		if titleEl.Length() > 0 {
			title = strings.TrimSpace(titleEl.Text())
		} else {
			// Fallback to link text
			title = strings.TrimSpace(linkEl.Text())
		}
	} else {
		titleEl := book.Find(".a-link-normal.aok-block span div")
		if titleEl.Length() == 0 {
			return utils.Report("Can't find book title")
		}
		title = strings.TrimSpace(titleEl.Text())
	}

	if title == "" {
		return utils.Report("Empty book title")
	}

	bookThumbnail.Title = title
	return nil
}

func extractCover(bookThumbnail *models.BookThumbnail, book *goquery.Selection) error {
	// Try multiple selectors for finding the image
	imageEl := book.Find("img.s-image")
	if imageEl.Length() == 0 {
		imageEl = book.Find("img")
	}

	if imageEl.Length() == 0 {
		return utils.Report("Can't find book image")
	}

	cover, exists := imageEl.Attr("src")
	if !exists {
		// Try data-src as fallback
		cover, exists = imageEl.Attr("data-src")
		if !exists {
			return utils.Report("Can't find image src")
		}
	}

	bookThumbnail.Cover = utils.ResizeBookImage(cover, COVER_IMG_SIZE*3)
	return nil
}

func extractAuthors(bookThumbnail *models.BookThumbnail, book *goquery.Selection) {
	authorEl := book.Find(".a-size-small.a-link-child")
	if authorEl.Length() == 0 {
		bookThumbnail.Authors = []models.AuthorType{}
		return
	}

	authorNameEl := authorEl.Find("div")
	if authorNameEl.Length() == 0 {
		bookThumbnail.Authors = []models.AuthorType{}
		return
	}

	link := authorEl.AttrOr("href", "")
	name := strings.TrimSpace(authorNameEl.Text())
	id, _ := utils.ExtractID(link)

	bookThumbnail.Authors = []models.AuthorType{{
		ID:   id,
		Name: name,
		Link: link,
	}}
}

func extractAuthorsFromSearch(bookThumbnail *models.BookThumbnail, book *goquery.Selection) {
	var authors []models.AuthorType
	seenAuthors := make(map[string]bool) // Track seen author names to avoid duplicates

	// First, try to find authors in the structured format used by Amazon
	// Look for the pattern: "de AuthorName" or "by AuthorName" in spans
	book.Find("span.a-size-base").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())

		// Check if this span contains "de " (French) or "by " (English) indicating author
		if text == "de " || text == "by " {
			// The next span should contain the author name
			nextSpan := s.Next()
			if nextSpan.Length() > 0 && nextSpan.HasClass("a-size-base") {
				authorName := strings.TrimSpace(nextSpan.Text())
				// Make sure it's not a date or other metadata
				if authorName != "" && !strings.Contains(authorName, "|") &&
					!strings.Contains(authorName, "Edition") &&
					!strings.Contains(authorName, "2024") &&
					!strings.Contains(authorName, "2025") &&
					!seenAuthors[authorName] {
					authors = append(authors, models.AuthorType{
						ID:   "",
						Name: authorName,
						Link: "",
					})
					seenAuthors[authorName] = true
				}
			}
		}
	})

	// If no authors found with the structured approach, try looking for author links
	if len(authors) == 0 {
		book.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/e/") { // Author page links contain /e/
				name := strings.TrimSpace(s.Text())
				if name != "" && !strings.Contains(name, "rating") &&
					!strings.Contains(name, "Edition") &&
					!strings.Contains(name, "Format") &&
					!seenAuthors[name] {
					id, _ := utils.ExtractID(href)
					authors = append(authors, models.AuthorType{
						ID:   id,
						Name: name,
						Link: href,
					})
					seenAuthors[name] = true
				}
			}
		})
	}

	// If still no authors found, look in the div containing author info
	if len(authors) == 0 {
		book.Find("div.a-row").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())

			// Look for patterns like "Édition en Anglais | de AuthorName | date"
			if strings.Contains(text, " | de ") {
				parts := strings.Split(text, " | de ")
				if len(parts) > 1 {
					authorPart := parts[1]
					// Remove date and other info after the next |
					if strings.Contains(authorPart, " | ") {
						authorPart = strings.Split(authorPart, " | ")[0]
					}

					authorName := strings.TrimSpace(authorPart)
					if authorName != "" && !seenAuthors[authorName] {
						authors = append(authors, models.AuthorType{
							ID:   "",
							Name: authorName,
							Link: "",
						})
						seenAuthors[authorName] = true
					}
				}
			}
		})
	}

	bookThumbnail.Authors = authors
}

func extractRating(bookThumbnail *models.BookThumbnail, book *goquery.Selection) {
	// Look for rating in the icon alt text
	ratingEl := book.Find(".a-icon-alt")
	if ratingEl.Length() == 0 {
		bookThumbnail.Rating = -1
		return
	}

	ratingText := strings.TrimSpace(ratingEl.Text())

	// Handle different rating text formats
	// "4.8 out of 5 stars" or "4,8 sur 5 étoiles" etc.
	if strings.Contains(ratingText, "out of") || strings.Contains(ratingText, "sur") {
		parts := strings.Fields(ratingText)
		if len(parts) > 0 {
			ratingText = parts[0]
		}
	} else {
		// Fallback: take first part before space
		ratingText = strings.Split(ratingText, " ")[0]
	}

	ratingText = strings.ReplaceAll(ratingText, ",", ".")
	ratingText = strings.TrimSpace(ratingText)

	rating, err := strconv.ParseFloat(ratingText, 64)
	if err != nil {
		bookThumbnail.Rating = -1
		return
	}
	bookThumbnail.Rating = float32(rating)
}

func extractPageCount(document *goquery.Document, isSearch bool) int {
	var selector string
	if isSearch {
		selector = ".s-pagination-strip ul > *"
	} else {
		selector = ".a-pagination > *"
	}

	pageCount := 0
	document.Find(selector).Each(func(i int, s *goquery.Selection) {
		if (s.Is("li") || s.Is("span")) && s.Text() != "" {
			if pageNum, err := strconv.Atoi(s.Text()); err == nil && pageNum > pageCount {
				pageCount = pageNum
			}
		}
	})
	return pageCount
}

func fetchAndParseBooks(url string, isSearch bool) (*[]models.BookThumbnail, int, error) {
	content, status, err := utils.Fetch(url)
	if status == 404 {
		return nil, 0, utils.Report("Page not found")
	}
	if err != nil {
		return nil, 0, err
	}

	document, err := parseDocument(content)
	if err != nil {
		return nil, 0, err
	}

	// Different selectors for search vs regular pages
	var bookSelector string
	if isSearch {
		bookSelector = `[data-asin]:not([data-asin=""])[role=listitem]`
	} else {
		bookSelector = "[data-asin]"
	}

	books := document.Find(bookSelector)
	if books.Length() == 0 {
		return nil, 0, utils.Report("Can't find books")
	}

	var result []models.BookThumbnail
	books.Each(func(i int, book *goquery.Selection) {
		bookThumbnail, err := extractBookData(book, isSearch)
		if err == nil { // Only add if no error (skips audio books)
			result = append(result, bookThumbnail)
		}
	})

	pageCount := extractPageCount(document, isSearch)
	return &result, pageCount, nil
}

func loadFromCache(fileName string) (*[]models.BookThumbnail, int, error) {
	content, actualFileName, err := utils.ReadFile(fileName)
	if err != nil {
		return nil, 0, err
	}

	var cachedBooks []models.BookThumbnail
	err = utils.ParseJson(content, &cachedBooks)
	if err != nil {
		return nil, 0, utils.Report("Failed to parse cached content: " + err.Error())
	}

	// Extract pageCount from filename
	pageCount := 0
	parts := strings.Split(strings.TrimSuffix(actualFileName, ".json"), "-")
	if len(parts) > 1 {
		pageCountStr := parts[len(parts)-1]
		pageCount, _ = strconv.Atoi(pageCountStr)
	}

	return &cachedBooks, pageCount, nil
}

func saveToCache(fileName string, books []models.BookThumbnail) error {
	cacheContent, err := utils.ToJson(books)
	if err != nil {
		return utils.Report("Failed to convert books to JSON: " + err.Error())
	}

	return utils.WriteFile(fileName, cacheContent)
}

func FetchBooks(page int) (*[]models.BookThumbnail, int, error) {
	// Check cache first
	fileName := fmt.Sprintf("%s/%d-*.json", utils.CACHE_DIRECTORY, page)
	if utils.CacheValid(fileName, utils.CACHE_DURATION) {
		return loadFromCache(fileName)
	}

	// Fetch from web
	url := fmt.Sprintf("%s/gp/bestsellers/books/ref=zg_bs_pg_2_books?ie=UTF8&pg=%d", utils.AMAZON_URL, page)
	books, pageCount, err := fetchAndParseBooks(url, false)
	if err != nil {
		return nil, 0, utils.Report("Failed to fetch books: " + err.Error())
	}

	if len(*books) == 0 {
		return nil, 0, utils.Report("No books found on page " + strconv.Itoa(page))
	}

	// Save to cache
	cacheFileName := fmt.Sprintf("%s/%d-%d.json", utils.CACHE_DIRECTORY, page, pageCount)
	if err := saveToCache(cacheFileName, *books); err != nil {
		return nil, 0, utils.Report("Failed to write cache file: " + err.Error())
	}

	return books, pageCount, nil
}

func encodeSearchQuery(query string) string {
	return url.QueryEscape(query)
}

func normalizeQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", utils.Report("Query cannot be empty")
	}

	query = encodeSearchQuery(query)

	return query, nil
}

func SearchBooks(query string, page int) (*[]models.BookThumbnail, int, error) {
	normalizedQuery, err := normalizeQuery(query)
	if err != nil {
		return nil, 0, err
	}

	// Check cache first
	fileName := fmt.Sprintf("%s/search-%s-%d-*.json", utils.CACHE_DIRECTORY, normalizedQuery, page)
	if utils.CacheValid(fileName, utils.CACHE_DURATION) {
		return loadFromCache(fileName)
	}

	// Fetch from web
	url := fmt.Sprintf("%s/s?k=%s&i=stripbooks&crid=2G90TZW10SV2H&sprefix=%sstripbooks%%2C273&ref=nb_sb_ss_mvt-t11-ranker_2_9&page=%d",
		utils.AMAZON_URL, normalizedQuery, strings.Split(normalizedQuery, " ")[0], page)

	println("URL: " + url)

	books, pageCount, err := fetchAndParseBooks(url, true)
	if err != nil {
		return nil, 0, utils.Report("Failed to search books: " + err.Error())
	}

	if len(*books) == 0 {
		return nil, 0, utils.Report("No books found for query '" + query + "' on page " + strconv.Itoa(page))
	}

	// Save to cache
	cacheFileName := fmt.Sprintf("%s/search-%s-%d-%d.json", utils.CACHE_DIRECTORY, normalizedQuery, page, pageCount)
	if err := saveToCache(cacheFileName, *books); err != nil {
		return nil, 0, utils.Report("Failed to write cache file: " + err.Error())
	}

	return books, pageCount, nil
}
