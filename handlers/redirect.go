package handlers

import (
	"api/database" // Adjust this to your actual database package path
	"github.com/gofiber/fiber/v2"
)

func RedirectURL(c *fiber.Ctx) error {
	// 1. Grab the "2ljTW8" part from the URL
	key := c.Params("key")

	// 2. Look up the long URL in your database (you'll need to adapt this to your actual DB function)
	longURL, err := database.GetLongURL(key) 
	if err != nil || longURL == "" {
		return c.Status(404).SendString("Shortlink not found!")
	}

	// 3. (Optional but recommended) Increment your click counter here!
	// go database.IncrementClick(key) 

	// 4. Send the user to the destination
	return c.Redirect(longURL)
}