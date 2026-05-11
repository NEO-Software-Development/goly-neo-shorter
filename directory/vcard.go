package directory

import (
	"fmt"
	"strings"
)

// RenderVCard formats a published directory + its public links as RFC 6350 vCard 4.0.
// Only `public` links are included.
func RenderVCard(d *Directory) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCARD\r\n")
	b.WriteString("VERSION:4.0\r\n")
	b.WriteString("FN:" + vEscape(d.Name) + "\r\n")
	b.WriteString("ORG:" + vEscape(d.Name) + "\r\n")
	if d.Tagline != "" {
		b.WriteString("NOTE:" + vEscape(d.Tagline) + "\r\n")
	}

	for _, l := range d.Links {
		if l.Visibility != "public" {
			continue
		}
		switch l.Kind {
		case "phone":
			fmt.Fprintf(&b, "TEL;TYPE=voice:%s\r\n", vEscape(l.Value))
		case "whatsapp":
			fmt.Fprintf(&b, "TEL;TYPE=cell:%s\r\n", vEscape(l.Value))
		case "sms":
			fmt.Fprintf(&b, "TEL;TYPE=textphone:%s\r\n", vEscape(l.Value))
		case "email":
			fmt.Fprintf(&b, "EMAIL:%s\r\n", vEscape(l.Value))
		case "address":
			fmt.Fprintf(&b, "ADR:;;%s;;;;\r\n", vEscape(l.Value))
		case "website", "linkedin", "instagram", "x", "facebook", "youtube", "tiktok", "github", "telegram", "signal":
			fmt.Fprintf(&b, "URL:%s\r\n", vEscape(l.Value))
		}
	}

	b.WriteString("END:VCARD\r\n")
	return b.String()
}

// vEscape escapes the characters reserved by RFC 6350 §3.4.
func vEscape(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`;`, `\;`,
		`,`, `\,`,
		"\n", `\n`,
		"\r", "",
	)
	return r.Replace(s)
}
