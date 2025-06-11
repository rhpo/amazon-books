package main

import (
	"amazon/internal/database"
	"amazon/internal/routes"
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

	database.ConnectDB()
}

func main() {
	setup()

	app := fiber.New()
	PORT := os.Getenv("PORT")

	routes.RegisterBookRoutes(app.Group("/books"))
	routes.RegisterAuthorRoutes(app.Group("/authors"))
	routes.RegisterOrderRoutes(app.Group("/orders"))

	app.Get("/", func(client *fiber.Ctx) error {
		return client.Status(200).Type("html").SendString(`<h1>hello <span style="color: red">world!</span></h1>`)
	})

	log.Fatal(app.Listen(":" + PORT))
}
