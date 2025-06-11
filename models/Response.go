package models

type Response struct {
	Error string "json:\"error\""
	Code  string "json:\"code\""
	Data  any    "json:\"data\""
}

type BooksResponse struct {
	PageCount int             "json:\"pages\""
	Books     []BookThumbnail `json:"books"`
}
