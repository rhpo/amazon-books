package types

type AuthorType string

type Book struct {
	id, title string
	author    AuthorType
	rating    float32
}

type BookThumbnail struct {
	ID, Link, Title, Cover string

	Author AuthorType
	Rating float32
}
