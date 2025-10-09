package books

import (
	"amazon/models"
	"context"
	"fmt"
	"time"

	"github.com/nleeper/goment"
	gbooks "google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

var (
	srv    *gbooks.Service
	apiKey string
)

func maxCoverURL(volumeID string) string {
	return fmt.Sprintf("https://books.google.com/books/publisher/content/images/frontcover/%s?fife=w800-h1200&source=gbs_api", volumeID)
}

// Init sets up the Google Books service (optionally with an API key).
// You only need to call this once (e.g. at program startup).
func Init(key string) error {
	apiKey = key
	ctx := context.Background()

	var err error
	if key != "" {
		srv, err = gbooks.NewService(ctx, option.WithoutAuthentication())
	} else {
		srv, err = gbooks.NewService(ctx)
	}
	return err
}

func parseGoogleBooksDate(date string) goment.Goment {
	gm, _ := goment.New(date, "YYYY-MM-DD")
	if len(date) == 4 {
		gm, _ = goment.New(date, "YYYY")
	} else if len(date) == 7 {
		gm, _ = goment.New(date, "YYYY-MM")
	}
	return *gm
}

func CastVolumeToBook(book *gbooks.Volume) models.Book {
	res := models.Book{}
	info := book.VolumeInfo

	// ID
	res.ID = book.Id

	// Authors
	for _, gauthor := range info.Authors {
		res.Authors = append(res.Authors, models.AuthorType{
			ID:   "",
			Name: gauthor,
		})
	}

	// Title
	res.Title = info.Title

	// Cover
	res.Cover = maxCoverURL(book.Id)

	// Rating (not available)
	res.Rating = -1

	res.Publisher = info.Publisher

	// Date
	date := parseGoogleBooksDate(info.PublishedDate)
	res.PubDate = date.ToTime().Format(time.RFC3339)

	res.Language = info.Language

	res.Description = info.Description

	res.Pages = int(info.PageCount)

	res.IsGBook = true

	return res
}

func CastVolumeToBookThumbnail(book *gbooks.Volume) models.BookThumbnail {
	res := models.BookThumbnail{}
	info := book.VolumeInfo

	// ID
	res.ID = book.Id

	// Authors
	for _, gauthor := range info.Authors {
		res.Authors = append(res.Authors, models.AuthorType{
			ID:   "",
			Name: gauthor,
		})
	}

	// Title
	res.Title = info.Title

	// Cover
	res.Cover = maxCoverURL(book.Id)

	// Rating (not available)
	res.Rating = -1

	// Turn on GBook because it's a google book
	res.IsGBook = true

	return res
}

func CastVolumesToBookThumbnails(gbooks []*gbooks.Volume) *[]models.BookThumbnail {
	res := make([]models.BookThumbnail, 0)

	for _, gbook := range gbooks {
		res = append(res, CastVolumeToBookThumbnail(gbook))
	}

	return &res
}

func gBooksFilter[T []*gbooks.Volume](books T) T {
	filtered := make([]*gbooks.Volume, 0)

	for _, gbook := range books {
		links := gbook.VolumeInfo.ImageLinks
		if links == nil || (links.ExtraLarge == "" && links.Large == "" && links.Medium == "" && links.Small == "" && links.Thumbnail == "") {
			continue
		}

		date := parseGoogleBooksDate(gbook.VolumeInfo.PublishedDate)

		if date.Year() < 1920 {
			continue
		}

		filtered = append(filtered, gbook)
	}

	return filtered

}

// FetchGBooks searches for books by query.
func FetchGBooks(query string, max int) (*[]models.BookThumbnail, int, error) {
	if srv == nil {
		if err := Init(""); err != nil {
			return nil, 0, err
		}
	}

	call := srv.Volumes.List(query).MaxResults(40).OrderBy("relevance").PrintType("BOOKS")

	// if max > 0 {
	// 	call = call.MaxResults(int64(max)).Projection("FULL").OrderBy("relevance")
	// }

	resp, err := call.Do()
	if err != nil {
		return nil, 0, err
	}

	filtered := gBooksFilter(resp.Items)
	return CastVolumesToBookThumbnails(filtered), 1, nil
}

// FetchGBook retrieves one book by its ID.
func FetchGBook(id string) (*models.Book, string, error) {
	if srv == nil {
		if err := Init(""); err != nil {
			return nil, "error_google_api", err
		}
	}

	vol, err := srv.Volumes.Get(id).Do()
	if err != nil {
		return nil, "cannot_get_volume", err
	}

	book := CastVolumeToBook(vol)
	return &book, "", nil
}
