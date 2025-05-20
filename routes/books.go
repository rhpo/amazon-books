package routes

import (
	"amazon/scrapers"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// RegisterBookRoutes registers book-related routes to the provided router group.
func RegisterBookRoutes(router fiber.Router) {

	router.Get("/", func(c *fiber.Ctx) error {

		page, error := strconv.Atoi(c.Query("page"))
		if error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
				"error": "Missing 'page' query parameter",
			})
		}

		books, error := scrapers.FetchBooks(page)

		if error != nil {
			return error
		}

		return c.Status(fiber.StatusOK).JSON(books)
	})

	// domain.com/books/book
	router.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		book, error := scrapers.FetchBook(id)

		if error != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Book not found",
				"err":   error.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(book)
	})
}
