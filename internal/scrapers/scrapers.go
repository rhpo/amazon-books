package scrapers

import (
	"amazon/internal/scrapers/authors"
	"amazon/internal/scrapers/books"
)

// books
var FetchGBook = books.FetchGBook

var FetchBook = books.FetchBook
var FetchBooks = books.FetchBooks

var SearchBooks = books.SearchBooks
var LirekaSearchBooks = books.LirekaSearchBooks

// authors
var FetchAuthor = authors.FetchAuthor
