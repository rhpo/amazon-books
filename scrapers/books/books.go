package books

import (
	. "amazon/types"
	"amazon/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FetchBooks(page int) (*[]BookThumbnail, error) {
	var result []BookThumbnail = make([]BookThumbnail, 0)

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
		bookThumbnail := BookThumbnail{}
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
				bookThumbnail.Authors = []AuthorType{}
			} else {
				authorNameEl := authorEl.Find("div")
				if authorNameEl.Length() == 0 {
					return nil, utils.Report("Can't find book author name...")
				}

				var link string = authorEl.AttrOr("href", "")
				var name string = strings.TrimSpace(authorNameEl.Text())
				id, _ := utils.ExtractID(link)

				println("Link: " + link + " / id: " + id)

				bookThumbnail.Authors = []AuthorType{
					{
						ID:    id,
						Name:  name,
						Link:  link,
						Image: "", // From main page, there is no image available
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

	return &result, nil
}
