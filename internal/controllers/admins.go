package controllers

import (
	"amazon/internal/services"
	"amazon/internal/utils"
	"amazon/models"
	"fmt"
	"time"

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

// LoginAdmin handles the admin login process and returns a JWT token.
func (h AdminHandler) LoginAdmin(c *fiber.Ctx) error {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid request body: " + err.Error(),
			Code:  "invalid_request",
		})
	}

	admin, err := h.Service.Login(loginData.Username, loginData.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Response{
			Error: "Login failed: " + err.Error(),
			Code:  "login_failed",
		})
	}

	token, err := utils.GenerateJWT(fmt.Sprint(admin.ID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to generate token: " + err.Error(),
			Code:  "login_failed",
		})
	}

	// save token to the session
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Expires:  time.Now().Add(utils.TOKEN_EXPIRY), // Set expiration time
		Secure:   true,
		SameSite: "None",
	})

	admin.Password = ""

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Code: "success",
		Data: fiber.Map{
			"token": token,
			"admin": admin,
		},
	})
}

func (h AdminHandler) VerifyToken(c *fiber.Ctx) error {
	token := c.Cookies("token")

	if token == "" {
		fmt.Printf("Token not found!")

		return c.Status(fiber.StatusUnauthorized).JSON(models.Response{
			Error: "Unauthorized: token missing",
			Code:  "token_missing",
		})
	}

	var valid bool
	var adminID string
	var err error

	if valid, adminID, err = h.Service.IsValidAdmin(token); !valid || err != nil {

		fmt.Printf("Unauthorized: %s because %s", adminID, err)

		return c.Status(fiber.StatusUnauthorized).JSON(models.Response{
			Error: fmt.Sprintf("Unauthorized: %s", adminID),
			Code:  "token_invalid",
		})
	}

	admin, err := h.Service.GetAdminByID(adminID)

	if err != nil {
		fmt.Printf("Failed to retrieve admin: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: fmt.Sprintf("Failed to retrieve admin (not found with token: %s and id: %s)", token, adminID),
			Code:  "token_invalid",
		})
	}

	return c.Status(200).JSON(models.Response{
		Error: "",
		Code:  "success",
		Data: fiber.Map{
			"admin": admin,
		},
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

func (h AdminHandler) GetAllEmails(c *fiber.Ctx) error {
	var emailService services.EmailService = *services.NewEmailService()
	var emailArray []string // flattened array of emails

	emails, err := emailService.GetEmails()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to retrieve emails: " + err.Error(),
		})
	}

	for _, email := range emails {
		emailArray = append(emailArray, email.Email)
	}

	return c.Status(fiber.StatusOK).JSON(emailArray)
}
