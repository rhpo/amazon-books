package models

import (
	"encoding/json"
	"strings"
)

// APIResponse représente la réponse complète contenant l'objet "data".
type APIResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Image                string     `json:"image"`
	Price                float64    `json:"price"`
	ShippingFees         float64    `json:"shippingFees"`
	Currency             string     `json:"currency"`
	IsInStock            bool       `json:"isInStock"`
	EAN13                *string    `json:"EAN13,omitempty"`
	ASIN                 string     `json:"ASIN"`
	ISBN                 *string    `json:"ISBN,omitempty"`
	Brand                *string    `json:"brand,omitempty"`
	Category             Category   `json:"category"`
	Categories           []Category `json:"categories"`
	SiteURL              string     `json:"siteURL"`
	SiteHTML             *string    `json:"siteHtml,omitempty"`
	ProductHasVariations bool       `json:"productHasVariations"`
	StatusCode           int        `json:"statusCode"`
	IsFinished           *bool      `json:"isFinished,omitempty"`
	IsDead               *bool      `json:"isDead,omitempty"`
	HTMLLength           int        `json:"htmlLength"`
	CaptchaFound         bool       `json:"captchaFound"`
	IsHTMLPage           bool       `json:"isHtmlPage"`
	Host                 string     `json:"host"`
	Images               []string   `json:"images"`
	Seller               string     `json:"seller"`
	Prices               Prices     `json:"prices"`
	DeliveryDate         string     `json:"deliveryDate"`
	Reviews              Reviews    `json:"reviews"`
	OriginalPrice        *float64   `json:"originalPrice,omitempty"`
}

type Category struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Node string `json:"node"`
}

type Prices struct {
	New float64 `json:"new"`
}

type Reviews struct {
	Rating        *float64         `json:"rating,omitempty"`
	Count         *int             `json:"count,omitempty"`
	RatingByStars map[string]int   `json:"ratingByStars,omitempty"`
	List          []ReviewListItem `json:"list,omitempty"`
}

type ReviewListItem struct {
	Title    string   `json:"title"`
	Rating   int      `json:"rating"`
	Date     string   `json:"date"`
	Options  []string `json:"options,omitempty"`
	Verified bool     `json:"verified"`
	Text     string   `json:"text"`
}

// GetFrom fills the APIResponse from a JSON string.
func (a *APIResponse) GetFrom(jsonStr string) error {
	println("JSON String:")
	println(jsonStr)
	return json.Unmarshal([]byte(jsonStr), a)
}

// Migrate converts the APIResponse into a Book struct.
func (a *APIResponse) Migrate() Book {
	d := a.Data

	// Get average rating if available
	var avgRating float32
	if d.Reviews.Rating != nil {
		avgRating = float32(*d.Reviews.Rating)
	} else if len(d.Reviews.RatingByStars) > 0 {
		var total, count int
		for star, num := range d.Reviews.RatingByStars {
			switch star {
			case "1", "2", "3", "4", "5":
				total += int(star[0]-'0') * num
				count += num
			}
		}
		if count > 0 {
			avgRating = float32(total) / float32(count)
		}
	}

	return Book{
		ID:          d.ASIN,
		Pages:       0, // Amazon doesn’t provide page count here
		Title:       d.Title,
		Cover:       firstOrEmpty(d.Images),
		PubDate:     d.DeliveryDate, // best approximation available
		Language:    guessLanguage(d.Description, d.Category.Name),
		Publisher:   d.BrandOrFallback(),
		Description: cleanDesc(d.Description),

		Price:  float32(d.Prices.New),
		Rating: avgRating,

		Authors:   []AuthorType{}, // not provided in JSON
		Dimension: Dimension{},    // not provided in JSON
	}
}

// --- Helper methods ---

func (d Data) BrandOrFallback() string {
	if d.Brand != nil && *d.Brand != "" {
		return *d.Brand
	}
	if d.Seller != "" {
		return d.Seller
	}
	return "Unknown"
}

func firstOrEmpty(list []string) string {
	if len(list) > 0 {
		return list[0]
	}
	return ""
}

func cleanDesc(desc string) string {
	return strings.TrimSpace(strings.ReplaceAll(desc, "\n", " "))
}

func guessLanguage(desc, category string) string {
	desc = strings.ToLower(desc)
	if strings.Contains(desc, "anglais") || strings.Contains(category, "anglais") {
		return "English"
	}
	if strings.Contains(desc, "français") || strings.Contains(category, "français") {
		return "French"
	}
	return "Unknown"
}
