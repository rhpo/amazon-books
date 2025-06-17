package services

import (
	"amazon/internal/database"
	"amazon/internal/utils"
	"amazon/models"
	"errors"
	"log"
	"os"

	"gorm.io/gorm"
)

type AdminService struct {
}

func NewAdminService() *AdminService {
	return &AdminService{}
}

func (s *AdminService) CreateAdmin(admin *models.Admin) error {
	if admin.ID == 0 || admin.Name == "" {
		return errors.New("missing required fields")
	}

	// hash password with APP_SECRET
	admin.Password = utils.HashPassword(admin.Password, os.Getenv("APP_SECRET"))

	// Save admin
	return database.DB.Create(admin).Error
}

func (s *AdminService) GetAdminByID(id string) (*models.Admin, error) {
	var admin models.Admin
	if err := database.DB.First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return &admin, nil
}

func (s *AdminService) GetAllAdmins() ([]models.Admin, error) {
	var admins []models.Admin
	if err := database.DB.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (s *AdminService) Login(username, password string) (*models.Admin, error) {
	var admin models.Admin
	if err := database.DB.Where("username = ?", username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}

	log.Printf("Admin found: %s", admin.Username)
	// print given password and admin password
	log.Printf("Given password: %s, Admin password: %s", password, admin.Password)

	// !utils.CheckPasswordHash's parameters are:
	log.Printf("%s, %s, %s", password, admin.Password, os.Getenv("APP_SECRET"))

	if !utils.CheckPasswordHash(password, admin.Password, os.Getenv("APP_SECRET")) {
		return nil, errors.New("invalid password")
	}

	return &admin, nil
}

func (s *AdminService) DeleteAdmin(id string) error {
	var admin models.Admin
	if err := database.DB.First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("admin not found")
		}
		return err
	}

	if err := database.DB.Delete(&admin).Error; err != nil {
		return err
	}

	return nil
}

func (s *AdminService) AdminExists(id string) (bool, error) {
	var count int64
	if err := database.DB.Model(&models.Admin{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
