package types

type Author struct {
	Name  string
	About string
	Image string
}

type AuthorThumbnail struct {
	Name, Link, Image string
}

// type AuthorType string
type AuthorType AuthorThumbnail
