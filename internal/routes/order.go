package routes

import (
	"amazon/internal/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterOrderRoutes(router fiber.Router) {

	var orderHandler controllers.OrderHandler = controllers.NewOrderHandler()

	router.Post("/", orderHandler.PostOrder)
	router.Get("/:id", orderHandler.GetOrderByID)
	router.Get("/", orderHandler.GetAllOrders)
}
