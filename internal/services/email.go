package services

import (
	"amazon/internal/database"
	"amazon/models"
	"fmt"

	"gorm.io/gorm"
)

type EmailService struct {
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) AddEmail(email *models.Email) error {
	var existingEmail models.Email

	err := database.DB.Where("email = ?", email.Email).First(&existingEmail).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		// Real DB error
		return err
	}

	if existingEmail.Email != "" {
		// Email already exists
		return fmt.Errorf("email already exists")
	}

	// Create the email
	return database.DB.Create(email).Error
}

func (s *EmailService) GetEmails() ([]models.Email, error) {
	var emails []models.Email
	err := database.DB.Find(&emails).Error
	if err != nil {
		return nil, err
	}
	return emails, nil
}
