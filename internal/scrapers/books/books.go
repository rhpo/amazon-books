package books

import (
	"amazon/internal/utils"
	"amazon/models"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func fetchPage(url string) (*[]models.BookThumbnail, int, error) {
	var result []models.BookThumbnail = make([]models.BookThumbnail, 0)
	var pageCount int = 0
	content, error := utils.Fetch(url)

	if error != nil {
		return nil, 0, error
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	if error != nil {
		return nil, 0, error
	}

	println(content, " - ", url)
	books := document.Find("[data-asin]")
	if books.Length() == 0 {
		return nil, 0, utils.Report("Can't find books!")
	}

	for i := range books.Length() {
		bookThumbnail := models.BookThumbnail{}
		book := books.Eq(i)

		{ // ID
			bookID, exists := book.Attr("data-asin")
			if !exists {
				return nil, 0, utils.Report("Can't find book ID...")
			}
			bookThumbnail.ID = bookID
		}

		{ // Link
			bookLinkEl := book.Find(".a-link-normal.aok-block")
			if bookLinkEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book element (.a-link-normal.aok-block)")
			}

			bookLink, exists := bookLinkEl.Attr("href")
			if !exists {
				return nil, 0, utils.Report("Can't find book link (.a-link-normal.aok-block)")
			}
			bookThumbnail.Link = bookLink
		}

		{ // Title
			titleEl := book.Find(".a-link-normal.aok-block span div")
			if titleEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book title...")
			}
			bookThumbnail.Title = strings.TrimSpace(titleEl.Text())
		}

		{ // Cover
			imageEl := book.Find("img")
			if imageEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book image...")
			}
			bookThumbnail.Cover, _ = imageEl.Attr("src")
		}

		{ // Author
			authorEl := book.Find(".a-size-small.a-link-child")
			if authorEl.Length() == 0 {
				bookThumbnail.Authors = []models.AuthorType{}
			} else {
				authorNameEl := authorEl.Find("div")
				if authorNameEl.Length() == 0 {
					return nil, 0, utils.Report("Can't find book author name...")
				}

				var link string = authorEl.AttrOr("href", "")
				var name string = strings.TrimSpace(authorNameEl.Text())
				id, _ := utils.ExtractID(link)

				bookThumbnail.Authors = []models.AuthorType{
					{
						ID:   id,
						Name: name,
						Link: link,
					},
				}
			}
		}

		{ // Rating
			ratingEl := book.Find(".a-icon-alt")
			if ratingEl.Length() == 0 {
				bookThumbnail.Rating = -1
				result = append(result, bookThumbnail)
				continue
			}

			ratingText := strings.TrimSpace(ratingEl.Text())
			ratingText = strings.Split(ratingText, "out of")[0]
			ratingText = strings.ReplaceAll(ratingText, ",", ".")
			ratingText = strings.TrimSpace(ratingText)
			rating, err := strconv.ParseFloat(ratingText, 64)

			if err != nil {
				return nil, 0, utils.Report("Failed to parse rating value" + err.Error())
			}
			bookThumbnail.Rating = float32(rating)
		}

		result = append(result, bookThumbnail)
	}

	// determine the total number of pages
	pagination := document.Find(".a-pagination > *")
	if pagination.Length() > 0 {
		pagination.Each(func(i int, s *goquery.Selection) {
			if s.Is("li") && s.Text() != "" {
				pageNum, err := strconv.Atoi(s.Text())
				if err == nil && pageNum > pageCount {
					pageCount = pageNum
				}
			}
		})
	}

	return &result, pageCount, nil
}

func FetchBooks(page int) (*[]models.BookThumbnail, int, error) {
	var result []models.BookThumbnail = make([]models.BookThumbnail, 0)

	fileName := fmt.Sprintf("%s/%d-*.json", utils.CACHE_DIRECTORY, page)
	if utils.CacheValid(fileName, utils.CACHE_DURATION) {
		content, actualFileName, err := utils.ReadFile(fileName)
		if err != nil {
			return nil, 0, utils.Report("Failed to read cache file: " + err.Error())
		}

		var cachedBooks []models.BookThumbnail
		err = utils.ParseJson(content, &cachedBooks)

		if err != nil {
			return nil, 0, utils.Report("Failed to parse cached content: " + err.Error())
		}

		// Extract pageCount from the filename (last number after - and before .json)
		var pageCount int = 0
		parts := strings.Split(strings.TrimSuffix(actualFileName, ".json"), "-")
		if len(parts) > 1 {
			pageCountStr := parts[len(parts)-1]
			pageCount, _ = strconv.Atoi(pageCountStr)
		}

		return &cachedBooks, pageCount, nil
	}

	url := "https://www.amazon.com/best-sellers-books-Amazon/zgbs/books?pg=" + fmt.Sprint(page)

	books, pageCount, err := fetchPage(url)
	if err != nil {
		return nil, 0, utils.Report("Failed to fetch books: " + err.Error())
	}

	// Add fetched books to the result
	result = append(result, *books...)
	if len(result) == 0 {
		return nil, 0, utils.Report("No books found on page " + strconv.Itoa(page))
	}

	{ // save to cache
		cacheContent, err := utils.ToJson(result)
		if err != nil {
			return nil, 0, utils.Report("Failed to convert books to JSON: " + err.Error())
		}

		newfileName := fmt.Sprintf("%s/%d-%d.json", utils.CACHE_DIRECTORY, page, pageCount)
		err = utils.WriteFile(newfileName, cacheContent)
		if err != nil {
			return nil, 0, utils.Report("Failed to write cache file: " + err.Error())
		}
	}

	return &result, pageCount, nil
}

// SEARCH BOOKS ---

func fetchSearchPage(url string) (*[]models.BookThumbnail, int, error) {
	var result []models.BookThumbnail = make([]models.BookThumbnail, 0)
	var pageCount int = 0

	content, error := utils.Fetch(url)

	if error != nil {
		return nil, 0, error
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	log.Println("Fetched content for URL:", url)

	if error != nil {
		return nil, 0, utils.Report("Failed to parse document: " + error.Error())
	}

	books := document.Find(`[data-asin]:not([data-asin=""])`)
	if books.Length() == 0 {
		return nil, 0, utils.Report("Can't find books!")
	}

	for i := range books.Length() {
		bookThumbnail := models.BookThumbnail{}
		book := books.Eq(i)

		{ // ID
			bookID, exists := book.Attr("data-asin")
			if !exists {
				return nil, 0, utils.Report("Can't find book ID...")
			}
			bookThumbnail.ID = bookID
		}

		{ // Link
			bookLinkEl := book.Find(".a-link-normal")
			if bookLinkEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book element (.a-link-normal)")
			}

			bookLink, exists := bookLinkEl.Attr("href")
			if !exists {
				return nil, 0, utils.Report("Can't find book link (.a-link-normal)")
			}
			bookThumbnail.Link = bookLink
		}

		{ // Title
			titleEl := book.Find(".a-link-normal h2 span")
			if titleEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book title...")
			}
			bookThumbnail.Title = strings.TrimSpace(titleEl.Text())
		}

		{ // Cover
			imageEl := book.Find("img")
			if imageEl.Length() == 0 {
				return nil, 0, utils.Report("Can't find book image...")
			}
			bookThumbnail.Cover, _ = imageEl.Attr("src")
		}

		{ // Authors
			// 			authorWrapper := book.Find(".a-row:has(.a-size-base)").First()
			//
			// 			authors := []models.AuthorType{}
			// 			// Find all author links (with href) and plain author names (span.a-size-base)
			// 			authorWrapper.Find("a").Each(func(i int, s *goquery.Selection) {
			// 				goHref, exists := s.Attr("href")
			// 				name := strings.TrimSpace(s.Text())
			// 				if exists && name != "" {
			// 					id, _ := utils.ExtractID(goHref)
			// 					authors = append(authors, models.AuthorType{
			// 						ID:   id,
			// 						Name: name,
			// 						Link: goHref,
			// 					})
			// 				}
			// 			})
			// 			authorWrapper.Find(".a-size-base").Each(func(i int, s *goquery.Selection) {
			// 				name := strings.TrimSpace(s.Text())
			// 				if name != "" {
			// 					authors = append(authors, models.AuthorType{
			// 						ID:   "",
			// 						Name: name,
			// 						Link: "",
			// 					})
			// 				}
			// 			})

			// note: it's not always possible to find authors in search results
			bookThumbnail.Authors = nil // []models.AuthorType{}
		}

		{ // Rating
			ratingEl := book.Find(".a-icon-alt")
			if ratingEl.Length() == 0 {
				bookThumbnail.Rating = -1
				result = append(result, bookThumbnail)
				continue
			}

			ratingText := strings.TrimSpace(ratingEl.Text())
			ratingText = strings.Split(ratingText, "out of")[0]
			ratingText = strings.ReplaceAll(ratingText, ",", ".")
			ratingText = strings.TrimSpace(ratingText)
			rating, err := strconv.ParseFloat(ratingText, 64)

			if err != nil {
				return nil, 0, utils.Report("Failed to parse rating value" + err.Error())
			}
			bookThumbnail.Rating = float32(rating)
		}

		result = append(result, bookThumbnail)
	}

	// determine the total number of pages
	pagination := document.Find(".s-pagination-strip ul > *")
	if pagination.Length() > 0 {
		pagination.Each(func(i int, s *goquery.Selection) {
			if (s.Is("li") || s.Is("span")) && s.Text() != "" {
				pageNum, err := strconv.Atoi(s.Text())
				if err == nil && pageNum > pageCount {
					pageCount = pageNum
				}
			}
		})
	}
	return &result, pageCount, nil
}

func SearchBooks(query string, page int) (*[]models.BookThumbnail, int, error) {
	var result []models.BookThumbnail = make([]models.BookThumbnail, 0)

	// Goal: Normalize the query by removing spaces and converting to lowercase
	query = strings.TrimSpace(query)
	query = strings.ToLower(query)
	query = strings.ReplaceAll(query, " ", "-")
	query = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return -1
	}, query)

	if query == "" {
		return nil, 0, utils.Report("Query cannot be empty")
	}

	fileName := fmt.Sprintf("%s/search-%s-%d-*.json", utils.CACHE_DIRECTORY, query, page)
	if utils.CacheValid(fileName, utils.CACHE_DURATION) {
		content, actualFileName, err := utils.ReadFile(fileName)
		if err != nil {
			return nil, 0, utils.Report("Failed to read cache file: " + err.Error())
		}

		var cachedBooks []models.BookThumbnail
		err = utils.ParseJson(content, &cachedBooks)

		if err != nil {
			return nil, 0, utils.Report("Failed to parse cached content: " + err.Error())
		}

		// Extract pageCount from the filename (last number after - and before .json)
		var pageCount int = 0
		parts := strings.Split(strings.TrimSuffix(actualFileName, ".json"), "-")
		if len(parts) > 2 {
			pageCountStr := parts[len(parts)-1]
			pageCount, _ = strconv.Atoi(pageCountStr)
		}

		return &cachedBooks, pageCount, nil
	}

	url := "https://www.amazon.com/s?k=" + query + "&i=stripbooks-intl-ship&page=" + strconv.Itoa(page)

	books, pageCount, err := fetchSearchPage(url)
	if err != nil {
		return nil, 0, utils.Report("Failed to search books: " + err.Error())
	}

	result = append(result, *books...)
	if len(result) == 0 {
		return nil, 0, utils.Report("No books found for query '" + query + "' on page " + strconv.Itoa(page))
	}

	{ // save to cache
		cacheContent, err := utils.ToJson(result)
		if err != nil {
			return nil, 0, utils.Report("Failed to convert books to JSON: " + err.Error())
		}

		newfileName := fmt.Sprintf("%s/search-%s-%d-%d.json", utils.CACHE_DIRECTORY, query, page, pageCount)
		err = utils.WriteFile(newfileName, cacheContent)
		if err != nil {
			return nil, 0, utils.Report("Failed to write cache file: " + err.Error())
		}
	}

	return &result, pageCount, nil
}
