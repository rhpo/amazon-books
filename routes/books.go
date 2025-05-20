package routes

import (
	"amazon/scrapers"
	"amazon/types"

	"github.com/gofiber/fiber/v2"
)

// RegisterBookRoutes registers book-related routes to the provided router group.
func RegisterBookRoutes(router fiber.Router) {
	router.Get("/", func(c *fiber.Ctx) error {

		id := c.Query("page")
		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
				"error": "Missing 'page' query parameter",
			})
		}
		books, error := scrapers.FetchBooks()

		if error != nil {
			return error
		}

		return c.Status(fiber.StatusOK).JSON(books)
	})

	router.Get("/book", func(c *fiber.Ctx) error {
		id := c.Query("id")
		book := types.Book{}

		return c.Status(fiber.StatusOK).JSON(book)
	})
}
