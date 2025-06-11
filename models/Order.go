package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	BookID  string `json:"book_id"`
}
