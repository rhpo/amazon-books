package books

import (
	"amazon/internal/utils"
	"amazon/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FetchBooks(page int) (*[]models.BookThumbnail, error) {
	var result []models.BookThumbnail = make([]models.BookThumbnail, 0)

	filename := fmt.Sprintf("%s/%d.json", utils.CACHE_DIRECTORY, page)
	if utils.CacheValid(filename, utils.CACHE_DURATION) {
		content, err := utils.ReadFile(filename)
		if err != nil {
			return nil, utils.Report("Failed to read cache file: " + err.Error())
		}

		var cachedBooks []models.BookThumbnail
		err = utils.ParseJson(content, &cachedBooks)

		if err != nil {
			return nil, utils.Report("Failed to parse cached content: " + err.Error())
		}
		return &cachedBooks, nil
	}

	url := "https://www.amazon.com/best-sellers-books-Amazon/zgbs/books?pg=" + fmt.Sprint(page)
	content, error := utils.Fetch(url)

	if error != nil {
		return nil, error
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	if error != nil {
		return nil, error
	}

	books := document.Find("[data-asin]")
	if books.Length() == 0 {
		return nil, utils.Report("Can't find books!")
	}

	for i := range books.Length() {
		bookThumbnail := models.BookThumbnail{}
		book := books.Eq(i)

		{ // ID
			bookID, exists := book.Attr("data-asin")
			if !exists {
				return nil, utils.Report("Can't find book ID...")
			}
			bookThumbnail.ID = bookID
		}

		{ // Link
			bookLinkEl := book.Find(".a-link-normal.aok-block")
			if bookLinkEl.Length() == 0 {
				return nil, utils.Report("Can't find book element (.a-link-normal.aok-block)")
			}

			bookLink, exists := bookLinkEl.Attr("href")
			if !exists {
				return nil, utils.Report("Can't find book link (.a-link-normal.aok-block)")
			}
			bookThumbnail.Link = bookLink
		}

		{ // Title
			titleEl := book.Find(".a-link-normal.aok-block span div")
			if titleEl.Length() == 0 {
				return nil, utils.Report("Can't find book title...")
			}
			bookThumbnail.Title = strings.TrimSpace(titleEl.Text())
		}

		{ // Cover
			imageEl := book.Find("img")
			if imageEl.Length() == 0 {
				return nil, utils.Report("Can't find book image...")
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
					return nil, utils.Report("Can't find book author name...")
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
				return nil, utils.Report("Failed to parse rating value" + err.Error())
			}
			bookThumbnail.Rating = float32(rating)
		}

		result = append(result, bookThumbnail)
	}

	{ // save to cache
		cacheContent, err := utils.ToJson(result)
		if err != nil {
			return nil, utils.Report("Failed to convert books to JSON: " + err.Error())
		}

		err = utils.WriteFile(filename, cacheContent)
		if err != nil {
			return nil, utils.Report("Failed to write cache file: " + err.Error())
		}
	}

	return &result, nil
}
