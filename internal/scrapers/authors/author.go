package authors

import (
	"amazon/internal/utils"
	"amazon/models"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FetchAuthor(authorID string) (*models.Author, error) {
	var result models.Author = models.Author{}

	var url string = "https://www.amazon.com/stores/author/" + authorID + "/about?ccs_id=6ce47a19-8d2f-4cba-9089-a8cda76c5f9b"
	content, error := utils.Fetch(url)

	if error != nil {
		return nil, error
	}

	{ // Create File author.html containing content
		file, err := os.Create("author.html")

		if err != nil {
			return nil, utils.Report("Failed to create file: author.html")
		}

		defer file.Close()
		_, err = file.WriteString(content)

		if err != nil {
			return nil, utils.Report("Failed to write content to file: author.html")
		}

		if err := file.Close(); err != nil {
			return nil, utils.Report("Failed to close file: author.html")
		}
	}

	contentReader := strings.NewReader(content)
	document, error := goquery.NewDocumentFromReader(contentReader)

	if error != nil {
		return nil, error
	}

	{ // ID
		result.ID = authorID
	}

	{ // Name
		nameElement := document.Find("[data-csa-c-painter] h1")
		if nameElement.Length() == 0 {
			return nil, utils.Report("Can't find author title: ([data-csa-c-painter])...")
		}

		result.Name = nameElement.Text()
	}

	{ // Image
		imageElement := document.Find("[class^=\"AuthorBio__author-bio__author-picture\"] img")
		if imageElement.Length() == 0 {
			return nil, utils.Report("Can't find author image: ([data-csa-c-painter] img)...")
		}

		imgSrc, exists := imageElement.Attr("src")

		if !exists {
			return nil, utils.Report("Can't find author image src!!")
		}

		result.Image = imgSrc
	}

	{ // About
		aboutElement := document.Find("[class^=\"AuthorBio__author-bio__author-biography\"]")
		if aboutElement.Length() == 0 {
			return nil, utils.Report("Can't find author about section: ([class^=\"AuthorBio__author-bio__author-biography\"])...")
		}

		result.About = aboutElement.Text()
	}

	return &result, nil
}
