package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Name             string `json:"name"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	Address          string `json:"address"`
	SubscriptionCode string `json:"subscriptionCode"` // Add subscription code field

	Status string `json:"status"` // e.g., "pending", "shipped", "delivered"

	// Relationships
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	OrderID  uint   `json:"orderId"`  // Changed to match TypeScript interface
	ItemType string `json:"itemType"` // Changed to match TypeScript interface
	ItemID   string `json:"itemId"`   // Changed to match TypeScript interface
	Quantity int    `json:"quantity"`

	// Relationship back to order
	Order Order `json:"-" gorm:"foreignKey:OrderID"`
}
