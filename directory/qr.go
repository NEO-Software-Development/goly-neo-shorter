package directory

import (
	"errors"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQR returns a PNG-encoded QR for the given content at the requested
// pixel size. The content is always a URL we control (built from the canonical
// base URL + slug) so this function never accepts arbitrary user input.
func GenerateQR(content string, size int) ([]byte, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}
	if size < 64 {
		size = 64
	}
	if size > 1024 {
		size = 1024
	}
	return qrcode.Encode(content, qrcode.Medium, size)
}
