package controllers

import (
	"amazon/internal/services"
	"amazon/models"
	"amazon/notification"
	"fmt"
	"slices"

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

func isOneEmpty(elements ...string) bool {
	return slices.Contains(elements, "")
}

func (h OrderHandler) PostOrder(c *fiber.Ctx) error {
	var order models.Order

	fmt.Printf("%+v\n", c.Body())

	// Parse the JSON body directly into Order struct
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Invalid JSON format: " + err.Error(),
		})
	}

	// Basic validation
	if isOneEmpty(order.Address, order.Email, order.Phone, order.Name) {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Missing required fields (name, email, phone, address)",
		})
	}

	// Validate that order has at least one item
	if len(order.OrderItems) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Order must contain at least one item",
		})
	}

	// Validate each order item
	for i, item := range order.OrderItems {
		if item.ItemType != "book" && item.ItemType != "subscription" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Response{
				Error: "Invalid item type. Must be 'book' or 'subscription'",
			})
		}
		if item.ItemID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Response{
				Error: "Item ID is required for all order items",
			})
		}
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(models.Response{
				Error: "Quantity must be greater than 0 for all order items",
			})
		}

		// Clear the OrderID as it will be set by the database after order creation
		order.OrderItems[i].OrderID = 0
	}

	// 	file, fileErr := c.FormFile("screenshot")
	//
	// 	// Reject if missing
	// 	if fileErr != nil || file == nil || file.Filename == "" {
	// 		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
	// 			Error: "Screenshot is required.",
	// 		})
	// 	}
	//
	// 	// Validate extension
	// 	ext := strings.ToLower(filepath.Ext(file.Filename))
	// 	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
	// 		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
	// 			Error: "Only image files (.jpg, .jpeg, .png, .gif) are allowed.",
	// 		})
	// 	}
	//
	// 	// Validate size
	// 	if file.Size > utils.MAX_FILE_SIZE {
	// 		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
	// 			Error: "Image file is too large. Max size is 15MB.",
	// 		})
	// 	}
	//
	// 	// Save the file
	// 	fileName := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), file.Filename)
	// 	if err := c.SaveFile(file, fileName); err != nil {
	// 		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
	// 			Error: "Cannot save Screenshot: " + err.Error(),
	// 		})
	// 	}
	// 	order.Screenshot = fileName

	err := h.Service.CreateOrder(&order)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to create order: " + err.Error(),
		})
	}

	notification.Send("New Order - "+order.Name, fmt.Sprint("Livres: ", len(order.OrderItems)))

	// Reload the order with OrderItems to return complete data
	createdOrder, err := h.Service.GetOrderByID(fmt.Sprintf("%d", order.ID))
	if err != nil {
		// Return the order anyway, but OrderItems might be empty
		return c.Status(fiber.StatusCreated).JSON(order)
	}

	return c.Status(fiber.StatusCreated).JSON(createdOrder)
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

func (h OrderHandler) GetOrdersByEmail(c *fiber.Ctx) error {
	email := c.Params("email")

	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Email parameter is required",
		})
	}

	orders, err := h.Service.GetOrdersByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to retrieve orders: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

func (h OrderHandler) SetOrderStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	status := c.Params("status")

	if status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Error: "Status field is required",
		})
	}

	o := NewOrderHandler()
	err := o.Service.SetOrderStatus(id, status)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Code: "success",
		Data: fiber.Map{
			"message": "Order status updated successfully",
		},
	})
}

func (h OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	id := c.Params("id")

	// Call the service to delete the order
	err := h.Service.DeleteOrder(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Error: "Failed to delete order: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.Response{
		Code: "success",
		Data: fiber.Map{
			"message": "Order deleted successfully",
		},
	})
}
