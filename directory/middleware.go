package directory

import (
	"goly-app/auth"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// currentUserID extracts the authenticated user's ID from the request context.
// Returns 0 and false if no user is present (which should not happen under
// AuthMiddleware, but we fail closed regardless).
func currentUserID(c *fiber.Ctx) (uint, bool) {
	u, ok := c.Locals("user").(*auth.User)
	if !ok || u == nil {
		return 0, false
	}
	return u.ID, true
}

func parseUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// notFound is used for both "missing" and "not yours" so we never leak the
// existence of a directory that belongs to someone else.
func notFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
}

func badRequest(c *fiber.Ctx, err error) error {
	if v, ok := err.(*ValidationError); ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "validation_failed",
			"field": v.Field,
			"message": v.Message,
		})
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "bad_request",
		"message": err.Error(),
	})
}

func serverError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal_error"})
}
