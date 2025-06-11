package controllers

import (
	"amazon/internal/scrapers"

	"github.com/gofiber/fiber/v2"
)

type AuthorHandler struct{}

func NewAuthorHandler() AuthorHandler {
	return AuthorHandler{}
}

func (h AuthorHandler) GetAuthorByID(c *fiber.Ctx) error {

	// domain.com/authors/{author}
	id := c.Params("id")
	author, error := scrapers.FetchAuthor(id)

	if error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Author not found",
			"err":   error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(author)
}
