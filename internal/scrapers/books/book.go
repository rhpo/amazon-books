package books

import (
	"os"
	"strconv"
	"strings"
	"time"

	"amazon/internal/utils"
	"amazon/models"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/nleeper/goment"
)

func FetchBook(id string) (*models.Book, string, error) {
	var result models.Book = models.Book{}

	var url string = utils.AMAZON_URL + "/dp/" + id
	content, statusCode, error := utils.Fetch(url)

	println("Updated link:", url)

	if error != nil {
		return nil, "unhandled_error", error
	}

	if statusCode == 404 {
		return nil, "not_found", utils.Report("Book not found")
	}

	{ // Create File book.html containing content
		file, err := os.Create("book.html")

		if err != nil {
			return nil, "fs_error", utils.Report("Failed to create file: book.html")
		}

		defer file.Close()
		_, err = file.WriteString(content)

		if err != nil {
			return nil, "fs_error", utils.Report("Failed to write content to file: book.html")
		}

		if err := file.Close(); err != nil {
			return nil, "fs_error", utils.Report("Failed to close file: book.html")
		}
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	if error != nil {
		return nil, "parse_error", error
	}

	bookFrame := document.Find("#centerCol")
	// authorsFrame := document.Find("#leftCol")
	bookImageFrame := document.Find("#leftCol")
	bookPriceFrame := document.Find("#rightCol")
	if bookImageFrame.Length() == 0 {
		return nil, "server_error", utils.Report("Can't find book image frame (#leftCol)")
	}
	if bookPriceFrame.Length() == 0 {
		return nil, "server_error", utils.Report("Can't find book price frame (#rightCol)")
	}
	if bookFrame.Length() == 0 {
		return nil, "server_error", utils.Report("Can't find book frame (#centerCol)")
	}

	infoFrame := bookFrame.Find(".a-carousel")

	if infoFrame.Length() == 0 {
		return nil, "server_error", utils.Report("Can't find info frame (.a-carousel from #centerCol)")
	}

	// remove script elements from bookFrame
	bookFrame.Find("script").Remove()
	bookFrame.Find("link").Remove()
	bookFrame.Find("style").Remove()

	{ // ID
		result.ID = id
	}

	// if bookFrame contains Audible then don't add
	if utils.IsAudible(bookFrame.Text()) {
		return nil, "server_error", utils.Report("Book is an Audible/Audio book")
	}

	{ // Title
		titleEl := bookFrame.Find("#productTitle")
		if titleEl.Length() == 0 {
			return nil, "server_error", utils.Report("Can't find title (#productTitle)...")
		}

		result.Title = strings.TrimSpace(titleEl.Text())
	}

	{ // Description
		descriptionWrapper := bookFrame.Find(".a-expander-content")
		if descriptionWrapper == nil {
			return nil, "server_error", utils.Report("Can't find the Description wrapper (.a-expander-content)")
		}

		// remove links containing javascript
		descriptionWrapper.Find("a[href*='javascript:void']").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			s.ReplaceWithHtml(text)
		})

		descriptionWrapper.Find("a[href^='/']").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				absoluteURL := utils.AMAZON_URL + href
				s.SetAttr("href", absoluteURL)

			}
		})

		html, _ := descriptionWrapper.Html()
		markdown, err := htmltomarkdown.ConvertString(html)

		if err != nil {
			return nil, "server_error", utils.Report("Can't convert the description into markdown...")
		}

		result.Description = markdown
	}

	{ // Cover
		bookImage := bookImageFrame.Find("#imgTagWrapperId > img")
		if bookImage.Length() == 0 {
			return nil, "server_error", utils.Report("Image tag doesn't exist")
		}
		src, srcExists := bookImage.Attr("src")
		if !srcExists {
			return nil, "server_error", utils.Report("Image tag doesn't have a src attribute")
		}

		result.Cover = utils.ResizeBookImage(src, utils.COVER_IMG_SIZE*3)
	}

	{ // Pages
		pagesCount := 0
		pagesWrapper := bookFrame.Find("#rpi-attribute-book_details-fiona_pages > .rpi-attribute-value span")
		if pagesWrapper.Length() != 0 {

			pagesText := pagesWrapper.Text()
			if pagesText == "" {
				utils.Report("Pages count text is empty")
				// return nil, "server_error", utils.Report("Pages count text is empty")
			}

			parts := strings.Fields(pagesText)
			if len(parts) == 0 {
				utils.Report("Pages count text is not in expected format")
				// return nil, "server_error", utils.Report("Pages count text is not in expected format")
			}

			pageCount, err := strconv.Atoi(parts[0])
			pagesCount = pageCount
			if err != nil {
				utils.Report("Pages count is not a valid integer")
				// return nil, "server_error", utils.Report("Pages count is not a valid integer")
			}
		}

		result.Pages = pagesCount

	}

	{ // Language
		languageWrapper := bookFrame.Find("#rpi-attribute-language .rpi-attribute-value span")
		if languageWrapper.Length() == 0 {
			utils.Report("Cannot find the language wrapper (#rpi-attribute-language)")
			// return nil, "server_error", utils.Report("Cannot find the language wrapper (#rpi-attribute-language)")
		}

		languageText := languageWrapper.Text()
		if languageText == "" {
			utils.Report("Language text is empty")
			// return nil, "server_error", utils.Report("Language text is empty")
		}

		result.Language = languageText
	}

	{ // Publisher
		publisher := ""
		publisherWrapper := bookFrame.Find("#rpi-attribute-book_details-publisher .rpi-attribute-value span")
		if publisherWrapper.Length() != 0 {
			// return nil, "server_error", utils.Report("Cannot find the publisher wrapper (#rpi-attribute-book_details-publisher)")
			publisher = strings.TrimSpace(publisherWrapper.Text())

			if publisher == "" {
				utils.Report("Publisher text is empty")
				// return nil, "server_error", utils.Report("Publisher text is empty")
			}

		}

		result.Publisher = publisher
	}

	{ // Publication-Date
		pubdateWrapper := bookFrame.Find("#rpi-attribute-book_details-publication_date .rpi-attribute-value span")
		if pubdateWrapper.Length() == 0 {
			utils.Report("Cannot find the pubdate wrapper (#rpi-attribute-book_details-publication_date)")
			// return nil, "server_error", utils.Report("Cannot find the pubdate wrapper (#rpi-attribute-book_details-publication_date)")
		}

		pubdateText := pubdateWrapper.Text()
		if pubdateText == "" {
			utils.Report("Pubdate text is empty")
			// return nil, "server_error", utils.Report("Pubdate text is empty")
		}

		m, err := goment.New(pubdateText, "MMMM D, YYYY")
		if err != nil {
			// return nil, "server_error", utils.Report("Pubdate cannot be parsed.")

			utils.Report("Pubdate cannot be parsed.")
			result.PubDate = ""
		} else {
			result.PubDate = m.ToTime().Format(time.RFC3339)
		}

	}

	{ // Rating
		ratingWrapper := bookFrame.Find(".a-popover-trigger.a-declarative > .a-size-base.a-color-base")
		if ratingWrapper.Length() != 0 {

			ratingText := strings.TrimSpace(ratingWrapper.Text())

			if ratingText == "" {
				utils.Report("Rating text is empty")
				result.Rating = -1.0
				// return nil, "server_error", utils.Report("Rating text is empty")
			}

			ratingText = strings.ReplaceAll(ratingText, ",", ".")
			rating, err := strconv.ParseFloat(strings.TrimSpace(ratingText), 64)

			if err == nil {
				result.Rating = float32(rating)
			}

		} else {
			result.Rating = -1.0 // No rating found
		}

	}

	// 	{ // Author(s) OLD LOGIC
	// 		var authors []models.AuthorType = make([]models.AuthorType, 0)
	//
	// 		authorElements := authorsFrame.Find(".a-row.a-spacing-small.a-spacing-top-medium")
	//
	// 		if authorElements.Length() == 0 {
	// 			return nil, "server_error", utils.Report("Cannot find authors from (.a-row.a-spacing-small.a-spacing-top-medium)")
	// 		}
	//
	// 		authorElements.Each(func(i int, el *goquery.Selection) {
	// 			name := strings.TrimSpace(el.Find(".a-truncate").Text())
	// 			link, _ := el.Find(".a-column.a-span4 > a").Attr("href")
	// 			id, _ := utils.ExtractID(link)
	//
	// 			authors = append(authors, models.AuthorType{
	// 				ID:   id,
	// 				Name: name,
	// 				Link: link,
	// 			})
	// 		})
	//
	// 		result.Authors = authors
	// 	}

	{ // Author(s) NEW LOGIC
		var authors []models.AuthorType = make([]models.AuthorType, 0)

		authorElements := bookFrame.Find(".author a")

		if authorElements.Length() == 0 {
			utils.Report("Cannot find authors from (.a-row.a-spacing-small.a-spacing-top-medium)")
			// return nil, "server_error", utils.Report("Cannot find authors from (.a-row.a-spacing-small.a-spacing-top-medium)")
		}

		authorElements.Each(func(i int, el *goquery.Selection) {
			name := strings.TrimSpace(el.Text())
			link, _ := el.Attr("href")
			id, _ := utils.ExtractID(link)

			authors = append(authors, models.AuthorType{
				ID:   id,
				Name: name,
				Link: link,
			})
		})

		result.Authors = authors
	}

	{ // Dimentions
		var d models.Dimension = models.Dimension{}

		dimentionsWrapper := infoFrame.Find("li > #rpi-attribute-book_details-dimensions .rpi-attribute-value span")
		if dimentionsWrapper.Length() != 0 {
			dimentionsText := dimentionsWrapper.Text()
			if dimentionsText != "" {

				dimentionsText = strings.ReplaceAll(dimentionsText, " inches", "")
				dimentionsText = strings.ReplaceAll(dimentionsText, " cm", "")
				dims := strings.Split(dimentionsText, " x ")

				if len(dims) != 3 {

					utils.Report("Dimentions text is not in expected format")
					dims = []string{"0", "0", "0"}

					// return nil, "server_error", utils.Report("Dimentions text is not in expected format")
				}

				width, err1 := strconv.ParseFloat(strings.TrimSpace(dims[0]), 64)
				depth, err2 := strconv.ParseFloat(strings.TrimSpace(dims[1]), 64)
				height, err3 := strconv.ParseFloat(strings.TrimSpace(dims[2]), 64)

				if err1 != nil || err2 != nil || err3 != nil {
					// return nil, "server_error", utils.Report("Failed to parse dimention values")
					utils.Report("Failed to parse dimention values")
				}

				d.Width = width
				d.Depth = depth
				d.Height = height

			} else {
				utils.Report("Dimentions text is empty")
				// return nil, "server_error", utils.Report("Dimentions text is empty")
			}
		} else {
			// return nil, "server_error", utils.Report("Cannot find the dimentions wrapper (#rpi-attribute-book_details-dimensions)")
			utils.Report("Cannot find the dimentions wrapper (#rpi-attribute-book_details-dimensions)")
		}

		result.Dimension = d
	}

	{ // Price (NOW OPTIONAL)
		priceWrapper := bookPriceFrame.Find(".aok-offscreen")
		if priceWrapper.Length() == 0 {

			result.Price = 0

			// return nil, "server_error", utils.Report("Cannot find the price wrapper (#tmm-grid-swatch-HARDCOVER .slot-price span)")
			utils.Report("Cannot find the price wrapper (#tmm-grid-swatch-HARDCOVER .slot-price span)")
		}

		priceText := priceWrapper.Text()
		priceText = strings.ReplaceAll(priceText, "&nbsp;", "")
		priceText = strings.ReplaceAll(priceText, "€", "")
		priceText = strings.ReplaceAll(priceText, ",", ".")
		priceText = strings.Split(priceText, "     ")[0]
		priceText = strings.TrimSpace(priceText)
		if priceText == "" {
			// return nil, "server_error", utils.Report("Price text is empty")
			utils.Report("Price text is empty")

			result.Price = 0
		}

		priceText = strings.Split(priceText, " ")[0]

		priceText = strings.TrimSpace(
			strings.ReplaceAll(strings.ReplaceAll(priceText, "from", ""), "€", ""))

		price, err := strconv.ParseFloat(priceText, 32)
		if err != nil {

			// AS PRICE BECAME NOW OPTIONAL, WE JUST SET IT TO -1 IF IT CANNOT BE PARSED
			// return nil, "server_error", utils.Report("Failed to parse price value, " + err.Error())

			price = 0
		}

		result.Price = float32(price)
	}

	return &result, "", nil
}
