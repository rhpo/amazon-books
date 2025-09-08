package controllers

import (
	"amazon/internal/services"
	"amazon/internal/utils"
	"amazon/models"

	"github.com/gofiber/fiber/v2"
)

type EmailHandler struct {
	Service *services.EmailService
}

func NewEmailHandler(service *services.EmailService) *EmailHandler {
	return &EmailHandler{
		Service: service,
	}
}

func (h *EmailHandler) AddEmail(c *fiber.Ctx) error {

	// Parse the email from the request body
	var email models.Email
	if err := c.BodyParser(&email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid request body: " + err.Error(),
			Code:  "error",
		})
	}

	if !utils.IsValidEmail(email.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid email format",
			Code:  "error",
		})
	}

	err := h.Service.AddEmail(&email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: err.Error(),
			Code:  "exists",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.Response{
		Code: "done",
		Data: struct {
			Message string `json:"message"`
		}{
			Message: "Email added successfully",
		},
	})
}

func (h *EmailHandler) GetAllEmails(c *fiber.Ctx) error {
	emails, err := h.Service.GetEmails()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to retrieve emails: " + err.Error(),
			Code:  "error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Code: "success",
		Data: emails,
	})
}
