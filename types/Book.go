package types

import (
	"github.com/nleeper/goment"
)

type Book struct {
	ID, Pages int

	Title, Cover, Description,
	Language, Publisher string

	Rating, Price float32

	Authors   []AuthorType
	PubDate   goment.DateTime
	Dimension Dimention
}

type BookThumbnail struct {
	ID, Link, Title, Cover string

	Authors []AuthorType
	Rating  float32
}
