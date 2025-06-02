package routes

import (
	"amazon/scrapers"
	"strconv"

	"amazon/types"

	"github.com/gofiber/fiber/v2"
)

// RegisterBookRoutes registers book-related routes to the provided router group.
func RegisterBookRoutes(router fiber.Router) {

	router.Get("/", func(c *fiber.Ctx) error {

		page, error := strconv.Atoi(c.Query("page"))
		if error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(types.Response{
				Error: "Missing 'page' query parameter",
				Code:  "invalid_params",
				Data:  nil,
			})
		}

		books, error := scrapers.FetchBooks(page)

		if error != nil {
			return c.Status(fiber.StatusNotFound).JSON(types.Response{
				Error: "Can't find any book.",
				Code:  "empty_page",
				Data:  nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(types.Response{
			Error: "",
			Code:  "success",
			Data:  books,
		})

	})

	// domain.com/books/book
	router.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		book, error := scrapers.FetchBook(id)

		if error != nil {
			return c.Status(fiber.StatusNotFound).JSON(types.Response{
				Error: "Book not found: " + error.Error(),
				Code:  "book_not_found",
				Data:  nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(types.Response{
			Error: "",
			Code:  "success",
			Data:  book,
		})
	})
}
