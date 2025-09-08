package routes

import (
	"amazon/internal/controllers"
	"amazon/internal/services"

	"github.com/gofiber/fiber/v2"
)

func SetupEmailRoutes(app *fiber.App) {
	emailService := services.NewEmailService()
	emailHandler := controllers.NewEmailHandler(emailService)

	// Define the email routes
	app.Post("/emails", emailHandler.AddEmail) // Add email to list
}
