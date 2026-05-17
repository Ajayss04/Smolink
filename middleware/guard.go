package middleware

import (
	"api/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func AuthGuard(c *fiber.Ctx) error {
	// 1. Read the token directly from the "jwt" cookie
	tokenString := c.Cookies("jwt")

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized. Please login.",
		})
	}

	// 2. Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		// If the token is expired or invalid, clear the cookie and reject
		c.ClearCookie("jwt")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Session expired. Please login again.",
		})
	}

	// 3. Extract claims and pass the email to the next handler
	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)
	
	c.Request().Header.Set("Email", email)
	
	return c.Next()
}