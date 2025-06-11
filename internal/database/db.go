package database

import (
	"amazon/internal/utils"
	"amazon/models"

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
}
