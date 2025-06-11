package controllers

import (
	"amazon/internal/scrapers"
	"amazon/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BookHandler struct{}

func NewBookHandler() *BookHandler {
	return &BookHandler{}
}

func (h *BookHandler) GetBooks(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Missing or invalid 'page' query parameter",
			Code:  "invalid_params",
			Data:  nil,
		})
	}

	books, err := scrapers.FetchBooks(page)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Can't find any book.",
			Code:  "empty_page",
			Data:  nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data: models.BooksResponse{
			PageCount: 1,
			Books:     *books,
		},
	})
}

func (h *BookHandler) GetBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	book, err := scrapers.FetchBook(id)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Book not found: " + err.Error(),
			Code:  "book_not_found",
			Data:  nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data:  book,
	})
}
