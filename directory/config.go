package directory

import (
	"os"
	"strings"
)

// PublicBaseURL is the canonical web origin used when generating QR codes and
// any client-facing absolute URLs. Encoding a server-controlled URL is what
// keeps the QR endpoint free of SSRF / phishing risk.
//
// Defaults to http://localhost:3000 for local development; set PUBLIC_BASE_URL
// in production.
func PublicBaseURL() string {
	v := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	if v == "" {
		return "http://localhost:3000"
	}
	return v
}

func PublicURLForSlug(slug string) string {
	return PublicBaseURL() + "/c/" + slug
}
