package auth

import "github.com/gofiber/fiber/v2"

func SetupRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/register", Register)
	auth.Post("/login", Login)
}
