package main

import (
	"amazon/internal/database"
	"amazon/internal/routes"
	"amazon/internal/scrapers/books"
	"amazon/internal/utils"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

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

func init() {
	setup()

	gApiKey := os.Getenv("GOOGLE_API")
	books.Init(gApiKey)
}

// main is the entry point of the application, initializing the Fiber web server and setting up routes.
func main() {

	if false {
		scrapeCategories()
		return
	}

	app := fiber.New(fiber.Config{
		BodyLimit: utils.MAX_FILE_SIZE,
	})

	PORT := os.Getenv("PORT")

	// Get allowed origins from environment or use defaults
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		utils.Report("Can't find ALLOWED_ORIGINS in environment variables", true)
	}

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

// IsSaved checks if the bookID is present in the ids slice.
func IsSaved(ids []string, bookID string) bool {
	return slices.Contains(ids, bookID)
}

// scrapePage scrapes book information from Amazon based on a search query and page number.
//
// It retrieves a list of books matching the query, checks for existing cached book data, and fetches
// new book details if the cache is invalid or the book is not already saved. The function handles errors
// by panicking and prints the progress of the scraping process. It also includes a delay between requests
// to avoid overwhelming the server.
func scrapePage(query string, page int) {
	booksAmazon, _, err := books.SearchBooks(query, page)

	if err != nil {
		panic(err)
	}

	bookIDs := []string{}
	entries, err := os.ReadDir("books_cache/books")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if len(name) > 5 && name[len(name)-5:] == ".json" {
				bookIDs = append(bookIDs, name[:len(name)-5])
			}
		}
	}

	var savedNum int = 0

	for _, bookThumbnail := range *booksAmazon {

		fileName := fmt.Sprintf("%s/%s.json", utils.BOOKS_CACHE_DIRECTORY, bookThumbnail.ID)
		if !utils.CacheValid(fileName, 30*24*time.Hour) || !IsSaved(bookIDs, bookThumbnail.ID) {
			book, _, err := books.FetchBook(bookThumbnail.ID)

			if err != nil {
				panic(err)
			}

			println("Saved:", book.Title, "ID:", book.ID)
			savedNum++

			time.Sleep(1 * time.Second)
		} else {
			println("Already saved:", bookThumbnail.Title, "ID:", bookThumbnail.ID)

		}

		fmt.Printf("Progress: %.0f\n", float64(len(bookIDs)+savedNum)/float64(len(*booksAmazon))*100)

	}
}

func scrapeCategories() {

	const maxPages int = 2

	categories := [...]string{
		"Philosophie", "Développement Personnel", "Histoire", "Sciences", "Actualité, Politique et Société", "Adolescents", "Arts et photographie", "Bandes dessinées pour enfants", "Beaux livres", "Calendriers et Agendas", "Livres de cuisine, cuisine et vins", "Référence", "Droit", "Entreprise et Bourse", "Études supérieures", "Famille et bien-être", "Science-fiction et Fantasy", "Humour", "Informatique et internet", "Livres pour enfants", "Loisirs créatifs, décoration et maison", "Manga", "Nature et animaux", "Religions et Spiritualités", "Romance et littérature sentimentale", "Romans et littérature", "Mystère et suspense", "Santé, Forme et Diététique", "Sciences, Techniques et Médecine", "Sciences humaines", "Scolaire et Parascolaire", "Sports et loisirs", "Tourisme et voyages",
	}

	for i, category := range categories {
		categories[i] = utils.NormalizeQuery(categories[i])

		for i := range maxPages {
			println("Scraping category:", category, "/ page", i+1)
			scrapePage(categories[i], i+1)
			println("Finished scraping category:", category, "/ page", i+1)

		}
	}

}
