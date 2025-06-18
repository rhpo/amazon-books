package utils

import (
	"time"
)

const (
	AMAZON_URL = "https://www.amazon.com"

	CACHE_DIRECTORY = "books_cache"
	CACHE_DURATION  = 5 * time.Minute
	TOKEN_EXPIRY    = 72 * time.Hour // Duration for which the token is valid
)
