package routes

import (
	"amazon/internal/controllers"
	"amazon/internal/services"
	"amazon/internal/utils"

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

	// Example: verify token (replace with your actual verification logic)
	adminID, err := utils.ValidateJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: invalid token",
		})
	}

	// Example: check if admin still exists in DB (replace with your actual DB check)
	exists, err := AdminService.AdminExists(adminID)
	if err != nil || !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: admin not found",
		})
	}

	// Token and admin are valid, proceed
	return c.Next()
}

func RegisterAdminRoutes(router fiber.Router) {

	var adminHandler controllers.AdminHandler = controllers.NewAdminHandler()

	router.Post("/", adminHandler.CreateAdmin)
	router.Post("/login", adminHandler.LoginAdmin)

	router.Get("/:id", RequireAdminLogin, adminHandler.GetAdminByID)
	router.Get("/", RequireAdminLogin, adminHandler.GetAllAdmins)
	router.Delete("/:id", RequireAdminLogin, adminHandler.DeleteAdmin)
}
