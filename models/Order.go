package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	// gorm.Model
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	BookID  string `json:"book_id"`
}
