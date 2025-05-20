package scrapers

import (
	. "amazon/types"
	"amazon/utils"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FetchBooks(page int) (*[]BookThumbnail, error) {
	var result []BookThumbnail = make([]BookThumbnail, 0)

	url := "https://www.amazon.com/best-sellers-books-Amazon/zgbs/books?pg=" + (string)(page)
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
	books.Each(func(i int, book *goquery.Selection) {
		result[i] = BookThumbnail{}

	})

	return &result, nil
}
