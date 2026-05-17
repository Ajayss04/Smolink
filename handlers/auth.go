package handlers

import (
	"api/config"
	"api/database"
	"api/types"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// Types
type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func Base(c *fiber.Ctx) error {
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "success",
		"message": "Welcome to smolink",
	})
}

func Login(c *fiber.Ctx) error {
	state := uuid.New().String()
	authConfig := config.AuthConf()
	url := authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	// Simply redirect the user directly to the Google login page
	return c.Redirect(url)
}

func GetJWT(c *fiber.Ctx) error {
	code := c.Query("code")
	authConf := config.AuthConf()
	ctx := context.Background()
	
	token, err := authConf.Exchange(ctx, code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot exchange code for token",
		})
	}
	
	client := authConf.Client(ctx, token)
	userData, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Cannot get user data")
	}
	defer userData.Body.Close()
	
	// Updated from ioutil.ReadAll to io.ReadAll
	body, err := io.ReadAll(userData.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Cannot read user data")
	}

	var userInfo UserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Cannot parse user data")
	}

	var user types.User
	user.Username = userInfo.Email
	user.Joined = time.Now().Format("2006-01-02 15:04:05")

	_, err = database.GetUser(userInfo.Email)
	if err != nil {
		database.RegisterUser(&user)
	}

	claims := jwt.MapClaims{
		"email": userInfo.Email,
	}
	
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Config("JWT_SECRET")))
	// Add this error check back in!
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Cannot create JWT token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,     // MUST be true for cross-domain cookies (requires HTTPS)
		SameSite: "None",   // Explicitly allows the cookie to be sent across different domains
		Path:     "/",
	})
	
	// Redirect to your frontend dashboard
	return c.Redirect("https://smolink-frontend.vercel.app")
}
