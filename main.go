package main

import (
	"amazon/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func setup() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	setup()

	app := fiber.New()
	PORT := os.Getenv("PORT")

	// ROUTE: Books
	routes.RegisterBookRoutes(app.Group("/books"))

	// ROUTE: Authors
	routes.RegisterAuthorRoutes(app.Group("/authors"))

	app.Get("/", func(client *fiber.Ctx) error {
		return client.SendFile("./index.html")
	})

	log.Fatal(app.Listen(":" + PORT))
}
