package auth

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Cookies("session_token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	session, err := GetSessionByToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if session.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "session expired",
		})
	}

	user, err := GetUserByID(session.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	c.Locals("user", user)

	return c.Next()
}

func AdminOnly(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
		})
	}

	return c.Next()
}
