package routes

import (
	"amazon/scrapers"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthorRoutes registers author-related routes to the provided router group.
func RegisterAuthorRoutes(router fiber.Router) {

	// domain.com/authors/{author}
	router.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		author, error := scrapers.FetchAuthor(id)

		if error != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Author not found",
				"err":   error.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(author)
	})
}
