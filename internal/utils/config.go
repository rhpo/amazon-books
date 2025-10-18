package utils

import (
	"time"
)

const (
	AMAZON_URL       = "https://www.amazon.fr"
	AMAZON_MEDIA_URL = "https://m.media-amazon.com"

	CACHE_DIRECTORY       = "books_cache"
	BOOKS_CACHE_DIRECTORY = "books_cache/books"
	CACHE_DURATION        = 5 * 12 * 30 * 24 * time.Hour // 5 Years
	MAX_FILE_SIZE         = 25 * 1024 * 1024   // 25 MB
	TOKEN_EXPIRY          = 3 * 24 * time.Hour // Duration for which the token is valid

	IS_DEVELOPMENT = true
)
