package scrapers

import (
	"fmt"
	"strings"

	. "amazon/types"
	"amazon/utils"

	"github.com/PuerkitoBio/goquery"
)

func FetchBook() (*[]BookThumbnail, error) {
	var result []BookThumbnail = make([]BookThumbnail, 0)

	const url string = "https://www.amazon.com/best-sellers-books-Amazon/zgbs/books"
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
		bookWrapper := BookThumbnail{}

		asin, exists := book.Attr("data-asin")
		if !exists {
			utils.Report("Can't find asin tag!")
			return nil, fmt.Errorf("Can't find asin tag!")
		}

		linkEl := book.Find(".a-link-normal")
		if linkEl.Length() <= 0 {
			utils.Report()
		}

		bookWrapper.ID = asin
	})

	return &result, nil
}
