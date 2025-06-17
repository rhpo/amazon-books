package database

import (
	"amazon/internal/utils"
	"amazon/models"
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	var driver gorm.Dialector = sqlite.Open("database.db")
	var config gorm.Config = gorm.Config{}

	database, err := gorm.Open(driver, &config)

	if err != nil {
		utils.Report("Failed to open database: "+err.Error(), true)
	}

	DB = database // Assign DB before migration
	Migrate()

	return DB
}

func Migrate() {
	DB.AutoMigrate(&models.Order{})
	DB.AutoMigrate(&models.Admin{})

	// add the default admin if it doesn't exist, checking by username
	username := os.Getenv("ADMIN_USERNAME")
	password := utils.HashPassword(os.Getenv("ADMIN_PASSWORD"), os.Getenv("APP_SECRET"))
	name := os.Getenv("ADMIN_NAME")

	var admin models.Admin
	result := DB.Where("username = ?", username).First(&admin)
	if result.Error == gorm.ErrRecordNotFound {
		admin = models.Admin{
			Username: username,
			Password: password,
			Name:     name,
		}
		DB.Create(&admin)
		if admin.ID == 0 {
			utils.Report("Default admin not created, please check the database connection and migration.", true)
		} else {
			utils.Report("Default admin created successfully with ID: "+fmt.Sprint(admin.ID), false)
		}
	}
}
