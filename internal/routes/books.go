package routes

import (
	"amazon/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterBookRoutes registers book-related routes to the provided router group.
func RegisterBookRoutes(router fiber.Router) {

	var bookHandler controllers.BookHandler = *controllers.NewBookHandler()

	// domain.com/books
	router.Get("/", bookHandler.GetBooks)

	// domain.com/books/search
	router.Get("/search", bookHandler.SearchBooks)

	// domain.com/books/book
	router.Get("/:id", bookHandler.GetBookByID)

}
