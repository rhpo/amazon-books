package models

type Book struct {
	ID    string `json:"id"`
	Pages int    `json:"pages"`

	Title       string `json:"title"`
	Cover       string `json:"cover"`
	PubDate     string `json:"publication_date"` // was goment.Time before, but decided to replace it with standard date
	Language    string `json:"language"`
	Publisher   string `json:"publisher"`
	Description string `json:"description"`

	Price  float32 `json:"price"`
	Rating float32 `json:"rating"`

	Authors   []AuthorType `json:"authors"`
	Dimension Dimension    `json:"dimensions"`
}

type BookThumbnail struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	Title string `json:"title"`
	Cover string `json:"cover"`

	Authors []AuthorType `json:"authors"`
	Rating  float32      `json:"rating"`
}
