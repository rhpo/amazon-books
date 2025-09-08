package main

import (
	"amazon/internal/database"
	"amazon/internal/routes"
	"amazon/internal/utils"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func setup() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.ConnectDB()

	// make uploads directory if not exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err = os.Mkdir("uploads", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func main() {
	setup()

	app := fiber.New(fiber.Config{
		BodyLimit: utils.MAX_FILE_SIZE,
	})

	PORT := os.Getenv("PORT")

	// Get allowed origins from environment or use defaults
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		utils.Report("Can't find ALLOWED_ORIGINS in environment variables", true)
	}

	println("Allowed Origins:", allowedOrigins)

	// Use Fiber's built-in CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,Cookie",
		AllowCredentials: true,
	}))

	routes.RegisterBookRoutes(app.Group("/books"))
	routes.RegisterAdminRoutes(app.Group("/admin"))
	routes.RegisterEmailRoutes(app.Group("/emails"))
	routes.RegisterAuthorRoutes(app.Group("/authors"))
	routes.RegisterOrderRoutes(app.Group("/orders"))

	app.Get("/", func(client *fiber.Ctx) error {
		return client.Status(200).Type("html").SendString(`<h1>Made by <a href="https://agency.codiha.com" style="color: royalblue">CODIHA</a> Agency.</h1>`)
	})

	log.Fatal(app.Listen(":" + PORT))
}
