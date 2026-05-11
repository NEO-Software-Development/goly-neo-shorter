package directory

import (
	"time"

	"goly-app/auth"
	mw "goly-app/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// SetupRoutes mounts the directory feature under /api/v1.
// Public read endpoints are wrapped in PrivacyLogger so visitor IPs/UAs are
// never persisted. Owner endpoints sit under AuthMiddleware and share the
// session-scoped rate limiter.
func SetupRoutes(router fiber.Router) {
	api := router.Group("/api/v1")

	publicReadLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: time.Minute,
	})

	publicRevealLimiter := limiter.New(limiter.Config{
		Max:        15,
		Expiration: time.Minute,
	})

	pub := api.Group("/c", mw.PrivacyLogger, publicReadLimiter)
	pub.Get("/:slug", PublicGetHandler)
	pub.Get("/:slug/qr.png", PublicQRHandler)
	pub.Get("/:slug/vcard", PublicVCardHandler)
	pub.Get("/:slug/links/:linkId/value", publicRevealLimiter, PublicLinkValueHandler)

	ownerLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: time.Minute,
	})

	dirs := api.Group("/directories", auth.AuthMiddleware, ownerLimiter)
	dirs.Get("/", ListDirectoriesHandler)
	dirs.Post("/", CreateDirectoryHandler)
	dirs.Get("/:id", GetDirectoryHandler)
	dirs.Patch("/:id", UpdateDirectoryHandler)
	dirs.Delete("/:id", DeleteDirectoryHandler)
	dirs.Post("/:id/publish", PublishDirectoryHandler)
	dirs.Get("/:id/audit", ListAuditHandler)
	dirs.Get("/:id/links", ListLinksHandler)
	dirs.Post("/:id/links", AddLinkHandler)
	dirs.Patch("/:id/links/:linkId", UpdateLinkHandler)
	dirs.Delete("/:id/links/:linkId", DeleteLinkHandler)
	dirs.Post("/:id/links/reorder", ReorderLinksHandler)
	dirs.Post("/:id/links/:linkId/verify/start", StartLinkVerificationHandler)
	dirs.Post("/:id/links/:linkId/verify/confirm", ConfirmLinkVerificationHandler)

	me := api.Group("/me", auth.AuthMiddleware, ownerLimiter)
	me.Get("/export", ExportMeHandler)
	me.Delete("/", DeleteMeHandler)
}
