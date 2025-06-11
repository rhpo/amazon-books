package routes

import (
	"amazon/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthorRoutes registers author-related routes to the provided router group.
func RegisterAuthorRoutes(router fiber.Router) {

	var authorHandler controllers.AuthorHandler = controllers.NewAuthorHandler()

	router.Get("/:id", authorHandler.GetAuthorByID)
}
