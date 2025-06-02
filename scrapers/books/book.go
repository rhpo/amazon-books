package books

import (
	"os"
	"strconv"
	"strings"

	. "amazon/types"
	"amazon/utils"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/nleeper/goment"
)

func FetchBook(id string) (*Book, error) {
	var result Book = Book{}

	var url string = "https://www.amazon.com/dp/" + id
	content, error := utils.Fetch(url)

	if error != nil {
		return nil, error
	}

	{ // Create File book.html containing content
		file, err := os.Create("book.html")

		if err != nil {
			return nil, utils.Report("Failed to create file: book.html")
		}

		defer file.Close()

		_, err = file.WriteString(content)

		if err != nil {
			return nil, utils.Report("Failed to write content to file: book.html")
		}

		if err := file.Close(); err != nil {
			return nil, utils.Report("Failed to close file: book.html")
		}
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	if error != nil {
		return nil, error
	}

	{ // ID
		result.ID = id
	}

	{ // Title
		titleEl := document.Find("#productTitle")
		if titleEl.Length() == 0 {
			return nil, utils.Report("Can't find title (#productTitle)...")
		}

		result.Title = strings.TrimSpace(titleEl.Text())
	}

	{ // Description
		descriptionWrapper := document.Find(".a-expander-content")
		if descriptionWrapper == nil {
			return nil, utils.Report("Can't find the Description wrapper (.a-expander-content)")
		}

		html, _ := descriptionWrapper.Html()
		markdown, err := htmltomarkdown.ConvertString(html)

		if err != nil {
			return nil, utils.Report("Can't convert the description into markdown...")
		}

		result.Description = markdown
	}

	{ // Cover
		bookImage := document.Find("#imgTagWrapperId > img")
		if bookImage.Length() == 0 {
			return nil, utils.Report("Image tag doesn't exist")
		}
		src, srcExists := bookImage.Attr("src")
		if !srcExists {
			return nil, utils.Report("Image tag doesn't have a src attribute")
		}

		result.Cover = src
	}

	{ // Pages
		pagesWrapper := document.Find("#rpi-attribute-book_details-fiona_pages > .rpi-attribute-value span")
		if pagesWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the pages count wrapper (#rpi-attribute-book_details-fiona_pages)")
		}

		pagesText := pagesWrapper.Text()
		if pagesText == "" {
			return nil, utils.Report("Pages count text is empty")
		}

		parts := strings.Fields(pagesText)
		if len(parts) == 0 {
			return nil, utils.Report("Pages count text is not in expected format")
		}

		pageCount, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, utils.Report("Pages count is not a valid integer")
		}

		result.Pages = pageCount

	}

	{ // Language
		languageWrapper := document.Find("#rpi-attribute-language .rpi-attribute-value span")
		if languageWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the language wrapper (#rpi-attribute-language)")
		}

		languageText := languageWrapper.Text()
		if languageText == "" {
			return nil, utils.Report("Language text is empty")
		}

		result.Language = languageText
	}

	{ // Publisher
		publisherWrapper := document.Find("#rpi-attribute-book_details-publisher .rpi-attribute-value span")
		if publisherWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the publisher wrapper (#rpi-attribute-book_details-publisher)")
		}

		publisherText := publisherWrapper.Text()
		if publisherText == "" {
			return nil, utils.Report("Publisher text is empty")
		}

		result.Publisher = publisherText
	}

	{ // Publication-Date
		pubdateWrapper := document.Find("#rpi-attribute-book_details-publication_date .rpi-attribute-value span")
		if pubdateWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the pubdate wrapper (#rpi-attribute-book_details-publication_date)")
		}

		pubdateText := pubdateWrapper.Text()
		if pubdateText == "" {
			return nil, utils.Report("Pubdate text is empty")
		}

		m, err := goment.New(pubdateText, "MMMM D, YYYY")
		if err != nil {
			return nil, utils.Report("Pubdate cannot be parsed.")
		}

		result.PubDate = m.ToDateTime()
	}

	{ // Rating
		ratingWrapper := document.Find(".cm-cr-review-stars-spacing-big span")
		if ratingWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the rating wrapper (.cm-cr-review-stars-spacing-big)")
		}

		ratingText := strings.TrimSpace(ratingWrapper.Text())
		ratingText = strings.Split(ratingText, "out of")[0]

		if ratingText == "" {
			return nil, utils.Report("Rating text is empty")
		}

		ratingText = strings.ReplaceAll(ratingText, ",", ".")
		rating, err := strconv.ParseFloat(strings.TrimSpace(ratingText), 64)

		if err != nil {
			return nil, utils.Report("Failed to parse rating value")
		}

		result.Rating = float32(rating)
	}

	{ // Author(s)
		var authors []AuthorType = make([]AuthorType, 0)

		authorElements := document.Find(".a-row.a-spacing-small.a-spacing-top-medium")

		if authorElements.Length() == 0 {
			return nil, utils.Report("Cannot find authors from (.cm-cr-review-stars-spacing-big)")
		}

		authorElements.Each(func(i int, el *goquery.Selection) {
			name := strings.TrimSpace(el.Find(".a-truncate").Text())
			link, _ := el.Find(".a-column.a-span4 > a").Attr("href")
			image, _ := el.Find(".a-column.a-span3").Find("img").Attr("src")
			id, _ := utils.ExtractID(link)

			authors = append(authors, AuthorType{
				ID:    id,
				Name:  name,
				Link:  link,
				Image: image,
			})
		})

		result.Authors = authors
	}

	{ // Dimentions
		var d Dimention = Dimention{}

		dimentionsWrapper := document.Find("#rpi-attribute-book_details-dimensions .rpi-attribute-value span")
		if dimentionsWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the dimentions wrapper (#rpi-attribute-book_details-dimensions)")
		}

		dimentionsText := dimentionsWrapper.Text()
		if dimentionsText == "" {
			return nil, utils.Report("Dimentions text is empty")
		}

		dimentionsText = strings.ReplaceAll(dimentionsText, " inches", "")
		dims := strings.Split(dimentionsText, " x ")

		if len(dims) != 3 {
			return nil, utils.Report("Dimentions text is not in expected format")
		}

		width, err1 := strconv.ParseFloat(strings.TrimSpace(dims[0]), 64)
		depth, err2 := strconv.ParseFloat(strings.TrimSpace(dims[1]), 64)
		height, err3 := strconv.ParseFloat(strings.TrimSpace(dims[2]), 64)

		if err1 != nil || err2 != nil || err3 != nil {
			return nil, utils.Report("Failed to parse dimention values")
		}

		d.Width = width
		d.Depth = depth
		d.Height = height

		result.Dimension = d
	}

	{ // Price
		priceWrapper := document.Find("#tmm-grid-swatch-HARDCOVER .slot-price span")
		if priceWrapper.Length() == 0 {
			return nil, utils.Report("Cannot find the price wrapper (#tmm-grid-swatch-HARDCOVER .slot-price span)")
		}

		priceText := priceWrapper.Text()
		if priceText == "" {
			return nil, utils.Report("Price text is empty")
		}

		priceText = strings.TrimSpace(
			strings.ReplaceAll(strings.ReplaceAll(priceText, "from", ""), "$", ""))

		price, err := strconv.ParseFloat(strings.TrimSpace(priceText), 32)
		if err != nil {
			return nil, utils.Report("Failed to parse price value, " + err.Error())
		}

		result.Price = float32(price)
	}

	return &result, nil
}
