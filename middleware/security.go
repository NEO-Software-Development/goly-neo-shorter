package middleware

import "github.com/gofiber/fiber/v2"

// SecurityHeaders attaches a baseline of hardening headers that apply to
// every response. Per-route handlers may override Referrer-Policy and
// X-Robots-Tag (the public directory handler tightens both).
func SecurityHeaders(c *fiber.Ctx) error {
	c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Referrer-Policy", "no-referrer-when-downgrade")
	c.Set("X-Frame-Options", "DENY")
	c.Set("Content-Security-Policy",
		"default-src 'self'; "+
			"img-src 'self' data: https:; "+
			"style-src 'self' 'unsafe-inline'; "+
			"script-src 'self'; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'")
	return c.Next()
}
