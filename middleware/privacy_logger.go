package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PrivacyLogger logs only method, path, status, and duration. No IP, no
// user-agent, no referer, no cookies. Use on public route groups so visitor
// requests leave no identifying trace on disk.
func PrivacyLogger(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	log.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start))
	return err
}
