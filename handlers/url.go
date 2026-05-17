package handlers

import (
	"api/database"
	"api/types"
	"fmt"
	"math/rand"

	"github.com/gofiber/fiber/v2"
)

// Helper function to generate a random 6-character short link
func generateShortKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func AddMapping(c *fiber.Ctx) error {
	username := c.Request().Header.Peek("Email")
	link := new(types.LinkDTO)

	// 1. Properly catch body parsing errors
	if err := c.BodyParser(link); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request payload",
		})
	}

	// 2. AUTO-GENERATE short URL if the user left it blank
	if link.ShortURL == "" {
		link.ShortURL = generateShortKey(6) 
	}

	// 3. Validate (now that ShortURL is guaranteed to exist)
	validationErr := link.Validate()
	if validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": validationErr.Error(),
		})
	}

	// 4. Store in MongoDB
	err := database.AddURL(link, string(username))
	if err != nil {
		// Type assertion to extract CustomError status code safely
		if customErr, ok := err.(*types.CustomError); ok {
			return c.Status(customErr.StatusCode()).JSON(fiber.Map{
				"status":  customErr.StatusCode(),
				"message": customErr.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  500,
			"message": "Internal server error",
		})
	}

	// 5. Cache in Redis
	if err := database.StoreMapping(link); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot store mapping in cache",
		})
	}

	fmt.Println("Mapping stored:", link.ShortURL)
	
	// 6. Return the generated short URL back to the client
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":    "success",
		"message":   "Mapping stored",
		"short_url": link.ShortURL,
	})
}

func GetAllShortLinks(c *fiber.Ctx) error {
	username := c.Request().Header.Peek("Email")
	mappings, err := database.GetUrlsByUser(string(username))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot get links",
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   mappings,
	})
}

func DeleteLink(c *fiber.Ctx) error {
	username := c.Request().Header.Peek("Email")
	shortURL := c.Params("shortURL")
	err := database.DeleteLink(shortURL, string(username))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Link deleted",
	})
}

func GetLinkById(c *fiber.Ctx) error {
	username := c.Request().Header.Peek("Email")
	shortURL := c.Params("shortURL")
	mapping, err := database.GetLinkInfo(shortURL, string(username))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot get link",
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   mapping,
	})
}

func RedirectToLongLink(c *fiber.Ctx) error {
    shortURL := c.Params("shortURL")
    
    // 1. Try Redis first
    longURL, err := database.GetLongURL(shortURL)
    
    if err != nil || longURL == "" {
        // 2. Fallback to MongoDB
        longURL, err = database.GetLongURLFromMongo(shortURL)
        if err != nil || longURL == "" {
            return c.Redirect("/404")
        }
        
        // 3. Restore to Redis so the next click is fast
        // (Re-creating the DTO just for Redis storage)
        database.StoreMapping(&types.LinkDTO{
            ShortURL: shortURL,
            LongURL:  longURL,
        })
    }
    
    // 4. Track click and redirect
    database.IncrementClickCount(shortURL)
    return c.Redirect(longURL)
}

func GetStats(c *fiber.Ctx) error {
    shortURL := c.Params("shortURL")
    count, err := database.GetClickCount(shortURL)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Cannot get stats",
        })
    }
    return c.JSON(fiber.Map{
        "status":     "success",
        "short_url":  shortURL,
        "click_count": count,
    })
}
