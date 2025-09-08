package routes

import (
	"amazon/internal/controllers"
	"amazon/internal/services"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func RequireAdminLogin(c *fiber.Ctx) error {

	var AdminService services.AdminService = *services.NewAdminService()

	token := c.Cookies("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: token missing",
		})
	}

	if valid, adminID, err := AdminService.IsValidAdmin(token); !valid || err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fmt.Sprintf("Unauthorized: %s", adminID),
		})
	}

	// Token and admin are valid, proceed
	return c.Next()
}

func RegisterAdminRoutes(router fiber.Router) {

	var adminHandler controllers.AdminHandler = controllers.NewAdminHandler()

	router.Post("/", RequireAdminLogin, adminHandler.CreateAdmin)
	router.Post("/login", adminHandler.LoginAdmin)

	router.Get("/verify", adminHandler.VerifyToken)

	router.Get("/:id", RequireAdminLogin, adminHandler.GetAdminByID)
	router.Get("/", RequireAdminLogin, adminHandler.GetAllAdmins)
	router.Delete("/:id", RequireAdminLogin, adminHandler.DeleteAdmin)

	router.Get("/emails", RequireAdminLogin, adminHandler.GetAllEmails)
}

func RegisterEmailRoutes(router fiber.Router) {
	emailService := services.NewEmailService()
	emailHandler := controllers.NewEmailHandler(emailService)

	// Define the email routes
	router.Post("/", emailHandler.AddEmail)                       // Add email to list
	router.Get("/", RequireAdminLogin, emailHandler.GetAllEmails) // Get all emails
}
