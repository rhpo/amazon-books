package utils

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
)

const DZD_TO_EUR = 260
const SERVICE_FEE = 800
const COVER_IMG_SIZE = 500

func ExtractID(url string) (string, error) {
	re := regexp.MustCompile(`/e/([A-Z0-9]+)(?:[/?]|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("ID not found in url: %s", url)
}

func Report(message string, mustRun ...any) error {
	fmt.Println(message)

	if len(mustRun) > 0 && mustRun[0].(bool) {
		panic(message)
	}

	return fmt.Errorf("%s", message)
}

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func Shuffle[T any](s []T) []T {
	for i := range s {
		j := rand.Intn(len(s))
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func FormatPrice(priceEur float32) float32 {
	if priceEur < 0 {
		return -1
	}

	priceDzd := float32(priceEur * DZD_TO_EUR)
	rounded := float32(int((priceDzd+4.99)/5) * 5) // round up to nearest 5
	return rounded + SERVICE_FEE
}

func ResizeBookImage(url string, size ...int) string {
	imageSize := COVER_IMG_SIZE
	if len(size) > 0 && size[0] > 0 {
		imageSize = size[0]
	}

	// Extract ID (between /I/ and .)
	parts := strings.Split(url, "/I/")
	if len(parts) < 2 {
		return url
	}
	id := strings.Split(parts[1], ".")[0]

	return fmt.Sprintf(AMAZON_MEDIA_URL+"/images/I/%s._SL%d_.jpg", id, imageSize)
}

func IsAudible(bookText string) bool {
	audible := strings.Contains(bookText, "audible")
	kindle := strings.Contains(bookText, "kindle")
	poche := strings.Contains(bookText, "poche")
	relie := strings.Contains(bookText, "relié")
	broche := strings.Contains(bookText, "broché")

	return audible && !kindle && !poche && !relie && !broche
}

func NormalizeQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	query = EncodeSearchQuery(query)

	return query
}

func EncodeSearchQuery(query string) string {
	return url.QueryEscape(query)
}
