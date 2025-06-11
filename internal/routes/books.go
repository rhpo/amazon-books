package routes

import (
	"amazon/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterBookRoutes registers book-related routes to the provided router group.
func RegisterBookRoutes(router fiber.Router) {

	var bookHandler controllers.BookHandler = *controllers.NewBookHandler()

	router.Get("/", func(c *fiber.Ctx) error {
		return bookHandler.GetBooks(c)
	})

	// domain.com/books/book
	router.Get("/:id", func(c *fiber.Ctx) error {
		return bookHandler.GetBookByID(c)
	})
}
