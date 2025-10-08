package controllers

import (
	"amazon/internal/scrapers"
	"amazon/internal/scrapers/books"
	"amazon/internal/utils"
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

	books, pageCount, err := scrapers.FetchBooks(page)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Can't find any book.",
			Code:  "empty_page",
			Data:  nil,
		})
	}

	// shuffled := utils.Shuffle(*books)
	// books = &shuffled

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data: models.BooksResponse{
			PageCount: pageCount,
			Books:     *books,
		},
	})
}

func (h *BookHandler) GetBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	book, errCode, err := scrapers.FetchBook(id)

	status := fiber.StatusInternalServerError
	if errCode == "not_found" {
		status = fiber.StatusNotFound
	}

	if err != nil {
		return c.Status(status).JSON(models.Response{
			Error: err.Error(),
			Code:  errCode,
			Data:  nil,
		})
	}

	// UPDATE: CLIENT ASKED TO NOT DISPLAY THE PRICE AND SEND IT AS AN EMAIL WITH HIS OWN FORMULA.
	// apply new formatted price
	// book.Price = utils.FormatPrice(book.Price)

	// upscale image for better quality
	book.Cover = utils.ResizeBookImage(book.Cover, 1000)

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data:  book,
	})
}

func (h *BookHandler) GetGBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	book, errCode, err := scrapers.FetchGBook(id)

	status := fiber.StatusInternalServerError
	if errCode == "not_found" {
		status = fiber.StatusNotFound
	}

	if err != nil {
		return c.Status(status).JSON(models.Response{
			Error: err.Error(),
			Code:  errCode,
			Data:  nil,
		})
	}

	// UPDATE: CLIENT ASKED TO NOT DISPLAY THE PRICE AND SEND IT AS AN EMAIL WITH HIS OWN FORMULA.
	// apply new formatted price
	// book.Price = utils.FormatPrice(book.Price)

	// upscale image for better quality
	book.Cover = utils.ResizeBookImage(book.Cover, 1000)

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data:  book,
	})
}

func (h *BookHandler) SearchBooks(c *fiber.Ctx) error {

	query := c.Query("query")
	_, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Missing or invalid 'page' query parameter",
			Code:  "invalid_params",
			Data:  nil,
		})
	}

	books, pageCount, err := books.FetchGBooks(query, 50)
	if err != nil {
		utils.Report("Failed to search books: " + err.Error())

		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "No search results for " + query,
			Code:  "empty_page",
			Data:  nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data: models.BooksResponse{
			PageCount: pageCount,
			Books:     *books,
		},
	})
}
