package main

import (
	"api/config"
	"api/database"
	"api/routes"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// app init
	app := fiber.New()

	// connect to database
	database.ConnectRedis()
	database.ConnectMongo()

	// Middleware
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://your-frontend-domain.vercel.app, http://127.0.0.1:5500", // Add your local testing & live frontend URLs
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH",
		AllowCredentials: true, // Critical for receiving and transmitting cookies
	}))

	// Routes
	routes.SetupRoutes(app)

	// Static files
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World")
	})
	app.Static("/404", "./public/404.html")
	app.Listen(config.Config("FIBER_PORT"))
	fmt.Println("Server is running on port", config.Config("FIBER_PORT"))
}
