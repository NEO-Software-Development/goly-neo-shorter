package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	loginLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
	})

	authGroup.Post("/register", loginLimiter, Register)
	authGroup.Post("/login", loginLimiter, Login)
	authGroup.Post("/logout", Logout)

	twofa := authGroup.Group("/2fa", AuthMiddleware)
	twofa.Post("/enroll", Enroll2FA)
	twofa.Post("/verify", Verify2FA)
	twofa.Post("/disable", Disable2FA)

	email := authGroup.Group("/email", AuthMiddleware)
	email.Post("/start", StartEmailVerification)
	email.Post("/confirm", ConfirmEmailVerification)

	sessions := authGroup.Group("/sessions", AuthMiddleware)
	sessions.Get("/", ListSessions)
	sessions.Delete("/:id", RevokeSession)
}
