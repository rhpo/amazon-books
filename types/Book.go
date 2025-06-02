package types

import (
	"github.com/nleeper/goment"
)

type Book struct {
	ID    string `json:"id"`
	Pages int    `json:"pages"`

	Title       string `json:"title"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Publisher   string `json:"publisher"`

	Rating float32 `json:"rating"`
	Price  float32 `json:"price"`

	Authors   []AuthorType    `json:"author"`
	PubDate   goment.DateTime `json:"publication_date"`
	Dimension Dimention       `json:"dimetions"`
}

type BookThumbnail struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	Title string `json:"title"`
	Cover string `json:"cover"`

	Authors []AuthorType `json:"authors"`
	Rating  float32      `json:"rating"`
}
