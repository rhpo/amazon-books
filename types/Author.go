package types

type Author struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	About string `json:"about"`
	Image string `json:"image"`
}

type AuthorThumbnail struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Link  string `json:"link"`
	Image string `json:"image"`
}

// type AuthorType string
type AuthorType AuthorThumbnail
