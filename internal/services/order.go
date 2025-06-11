package services

import (
	"amazon/internal/database"
	"amazon/models"
	"errors"

	"gorm.io/gorm"
)

type OrderService struct {
}

func NewOrderService() *OrderService {
	return &OrderService{}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	if order.BookID == "" || order.Name == "" {
		return errors.New("missing required fields")
	}

	// Check if the book exists
	// fetch and see if it returns 404 or not
	// Might not impement this logic...

	// Save order
	return database.DB.Create(order).Error
}

func (s *OrderService) GetOrderByID(id string) (*models.Order, error) {
	var order models.Order
	if err := database.DB.First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	if err := database.DB.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
