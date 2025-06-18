package controllers

import (
	"amazon/internal/services"
	"amazon/internal/utils"
	"amazon/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	Service *services.AdminService
}

func NewAdminHandler() AdminHandler {
	return AdminHandler{
		Service: services.NewAdminService(),
	}
}

func (h AdminHandler) CreateAdmin(c *fiber.Ctx) error {

	// Parse the admin from the request body
	var admin models.Admin
	if err := c.BodyParser(&admin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	err := h.Service.CreateAdmin(&admin)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to create admin: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(admin)
}

func (h AdminHandler) GetAdminByID(c *fiber.Ctx) error {
	id := c.Params("id")

	admin, err := h.Service.GetAdminByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Admin not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(admin)
}

func (h AdminHandler) GetAllAdmins(c *fiber.Ctx) error {
	admins, err := h.Service.GetAllAdmins()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to retrieve admins",
		})
	}

	return c.Status(fiber.StatusOK).JSON(admins)
}

func (h AdminHandler) LoginAdmin(c *fiber.Ctx) error {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	admin, err := h.Service.Login(loginData.Username, loginData.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Response{
			Error: "Login failed: " + err.Error(),
		})
	}

	token, err := utils.GenerateJWT(fmt.Sprint(admin.ID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to generate token: " + err.Error(),
		})
	}

	// save token to the session
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   true, // Set to true if using HTTPS
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
		"admin": admin,
	})
}

func (h AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.Service.DeleteAdmin(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Admin not found",
		})
	}

	return c.Status(fiber.StatusNoContent).SendString("Admin deleted successfully")
}
