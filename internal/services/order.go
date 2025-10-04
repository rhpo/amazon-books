package services

import (
	"amazon/internal/database"
	"amazon/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OrderService struct {
}

func NewOrderService() *OrderService {
	return &OrderService{}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	if order.Name == "" || order.Email == "" || order.Phone == "" || order.Address == "" {
		return errors.New("missing required fields (name, email, phone, address)")
	}

	order.Status = "new"

	// Save order
	return database.DB.Create(order).Error
}

func (s *OrderService) GetOrderByID(id string) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Preload("OrderItems").First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	if err := database.DB.Preload("OrderItems").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) DeleteOrder(id string) error {
	if err := database.DB.Delete(&models.Order{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetOrdersByEmail(email string) ([]models.Order, error) {
	var orders []models.Order
	if err := database.DB.Preload("OrderItems").Where("email = ?", email).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) SetOrderStatus(id string, status string) error {
	// Update order status
	if err := database.DB.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}
	return nil
}
