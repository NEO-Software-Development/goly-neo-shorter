package directory

import (
	"errors"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	slugRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,38}[a-z0-9])?$`)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
	hexRe  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

	reservedSlugs = map[string]struct{}{
		"admin": {}, "api": {}, "auth": {}, "c": {}, "r": {},
		"goly": {}, "login": {}, "register": {}, "logout": {},
		"me": {}, "static": {}, "assets": {}, "health": {},
		"qr": {}, "vcard": {}, "well-known": {}, "robots.txt": {},
		"favicon.ico": {}, "settings": {}, "directories": {}, "directory": {},
	}

	textPolicy = bluemonday.StrictPolicy()

	validKinds = map[string]struct{}{
		"phone": {}, "email": {}, "website": {}, "whatsapp": {},
		"telegram": {}, "signal": {}, "sms": {}, "linkedin": {},
		"instagram": {}, "x": {}, "facebook": {}, "youtube": {},
		"tiktok": {}, "github": {}, "address": {}, "custom": {},
	}

	sensitiveKinds = map[string]struct{}{
		"phone": {}, "email": {}, "whatsapp": {}, "sms": {},
	}

	urlKinds = map[string]struct{}{
		"website": {}, "linkedin": {}, "instagram": {}, "x": {},
		"facebook": {}, "youtube": {}, "tiktok": {}, "github": {},
		"telegram": {}, "signal": {},
	}
)

// ErrInvalid is returned with a stable code that handlers can map to 400.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func invalid(field, msg string) error { return &ValidationError{Field: field, Message: msg} }

func ValidateSlug(s string) error {
	if s == "" {
		return invalid("slug", "required")
	}
	if !slugRe.MatchString(s) {
		return invalid("slug", "must be 3-40 chars of [a-z0-9-], no leading/trailing hyphen")
	}
	if _, reserved := reservedSlugs[s]; reserved {
		return invalid("slug", "reserved")
	}
	return nil
}

func IsSensitiveKind(k string) bool { _, ok := sensitiveKinds[k]; return ok }

// ShouldHideValue reports whether the public list response should omit Value
// for this link. Falls back to "auto" when reveal mode is empty.
func (l ContactLink) ShouldHideValue() bool {
	switch l.RevealMode {
	case "inline":
		return false
	case "tap":
		return true
	default:
		return IsSensitiveKind(l.Kind)
	}
}

func ValidateDirectoryInput(d *Directory) error {
	d.Name = strings.TrimSpace(textPolicy.Sanitize(d.Name))
	d.Tagline = strings.TrimSpace(textPolicy.Sanitize(d.Tagline))

	if d.Name == "" {
		return invalid("name", "required")
	}
	if len(d.Name) > 120 {
		return invalid("name", "max 120 chars")
	}
	if len(d.Tagline) > 200 {
		return invalid("tagline", "max 200 chars")
	}
	if d.LogoURL != "" {
		if err := validateHTTPURL(d.LogoURL); err != nil {
			return invalid("logo_url", err.Error())
		}
	}
	if d.AccentColor != "" && !hexRe.MatchString(d.AccentColor) {
		return invalid("accent_color", "must be #rrggbb")
	}
	return nil
}

func ValidateContactLinkInput(l *ContactLink) error {
	l.Label = strings.TrimSpace(textPolicy.Sanitize(l.Label))
	l.Value = strings.TrimSpace(l.Value)

	if _, ok := validKinds[l.Kind]; !ok {
		return invalid("kind", "unknown contact kind")
	}
	if len(l.Label) > 60 {
		return invalid("label", "max 60 chars")
	}
	if l.Visibility == "" {
		l.Visibility = "public"
	}
	if l.Visibility != "public" && l.Visibility != "unlisted" {
		return invalid("visibility", "must be public or unlisted")
	}
	if l.RevealMode == "" {
		l.RevealMode = "auto"
	}
	if l.RevealMode != "auto" && l.RevealMode != "inline" && l.RevealMode != "tap" {
		return invalid("reveal_mode", "must be auto, inline, or tap")
	}
	if l.Value == "" {
		return invalid("value", "required")
	}

	switch l.Kind {
	case "phone", "whatsapp", "sms":
		if !e164Re.MatchString(l.Value) {
			return invalid("value", "must be E.164 (e.g. +12025550123)")
		}
	case "email":
		if _, err := mail.ParseAddress(l.Value); err != nil {
			return invalid("value", "must be a valid email address")
		}
	case "address":
		l.Value = textPolicy.Sanitize(l.Value)
		if len(l.Value) > 500 {
			return invalid("value", "max 500 chars")
		}
	case "custom":
		if l.Label == "" {
			return invalid("label", "required for custom links")
		}
		l.Value = textPolicy.Sanitize(l.Value)
		if len(l.Value) > 500 {
			return invalid("value", "max 500 chars")
		}
	default:
		if _, isURL := urlKinds[l.Kind]; isURL {
			if err := validateHTTPURL(l.Value); err != nil {
				return invalid("value", err.Error())
			}
		}
	}
	return nil
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("must be a valid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http and https schemes are allowed")
	}
	if u.Host == "" {
		return errors.New("URL must have a host")
	}
	return nil
}
