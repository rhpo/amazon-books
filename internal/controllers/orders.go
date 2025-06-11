package controllers

import (
	"amazon/internal/services"
	"amazon/models"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	Service *services.OrderService
}

func NewOrderHandler() OrderHandler {
	return OrderHandler{
		Service: services.NewOrderService(),
	}
}

func (h OrderHandler) PostOrder(c *fiber.Ctx) error {

	// Parse the order from the request body
	var order models.Order
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	err := h.Service.CreateOrder(&order)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to create order: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	id := c.Params("id")

	order, err := h.Service.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Error: "Order not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(order)
}

func (h OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.Service.GetAllOrders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to retrieve orders",
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}
