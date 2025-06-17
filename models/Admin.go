package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	// gorm.Model
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Name     string `json:"name"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}
