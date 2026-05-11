package directory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"goly-app/auth"
	"goly-app/goly/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// viewDebounce holds a tiny in-memory record of recent view bumps so reloads
// don't inflate the counter. Keys are hash(slug|ip), values are a timestamp.
// The map is never persisted to disk; entries are evicted lazily.
var (
	viewDebounce   = make(map[string]time.Time)
	viewDebounceMu sync.Mutex
	viewWindow     = 5 * time.Minute
)

func shouldCountView(slug, ip string) bool {
	if ip == "" {
		return true
	}
	h := sha256.Sum256([]byte(slug + "|" + ip))
	key := hex.EncodeToString(h[:8])

	viewDebounceMu.Lock()
	defer viewDebounceMu.Unlock()

	now := time.Now()
	if last, ok := viewDebounce[key]; ok && now.Sub(last) < viewWindow {
		return false
	}
	viewDebounce[key] = now

	if len(viewDebounce) > 4096 {
		for k, t := range viewDebounce {
			if now.Sub(t) > viewWindow {
				delete(viewDebounce, k)
			}
		}
	}
	return true
}

func respectsDNT(c *fiber.Ctx) bool {
	return c.Get("DNT") == "1" || c.Get("Sec-GPC") == "1"
}

func applyPrivacyHeaders(c *fiber.Ctx, indexable bool) {
	c.Set("Referrer-Policy", "no-referrer")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
	if !indexable {
		c.Set("X-Robots-Tag", "noindex, nofollow")
	}
}

// --- Owner endpoints ---------------------------------------------------------

type directoryInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Tagline     string `json:"tagline"`
	LogoURL     string `json:"logo_url"`
	AccentColor string `json:"accent_color"`
	IsPublished *bool  `json:"is_published"`
	IsIndexable *bool  `json:"is_indexable"`
}

func CreateDirectoryHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}

	var in directoryInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}

	d := &Directory{
		OwnerID:     ownerID,
		Name:        in.Name,
		Tagline:     in.Tagline,
		LogoURL:     in.LogoURL,
		AccentColor: in.AccentColor,
	}
	if err := ValidateDirectoryInput(d); err != nil {
		return badRequest(c, err)
	}

	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if slug != "" {
		if err := ValidateSlug(slug); err != nil {
			return badRequest(c, err)
		}
		exists, err := SlugExists(slug)
		if err != nil {
			return serverError(c)
		}
		if exists {
			slug = ""
		}
	}
	if slug == "" {
		for i := 0; i < 5; i++ {
			candidate := strings.ToLower(utils.RandomURL(8))
			if err := ValidateSlug(candidate); err != nil {
				continue
			}
			exists, err := SlugExists(candidate)
			if err != nil {
				return serverError(c)
			}
			if !exists {
				slug = candidate
				break
			}
		}
		if slug == "" {
			return serverError(c)
		}
	}
	d.Slug = slug

	if err := CreateDirectory(d); err != nil {
		return serverError(c)
	}
	WriteAudit(ownerID, d.ID, "directory.create", "slug="+d.Slug)
	return c.Status(fiber.StatusCreated).JSON(d)
}

func ListDirectoriesHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	out, err := ListDirectoriesByOwner(ownerID)
	if err != nil {
		return serverError(c)
	}
	return c.JSON(out)
}

func GetDirectoryHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	d, err := GetDirectoryForOwner(id, ownerID)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}
	return c.JSON(d)
}

func UpdateDirectoryHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	d, err := GetDirectoryForOwner(id, ownerID)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}

	var in directoryInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	if in.Name != "" {
		d.Name = in.Name
	}
	d.Tagline = in.Tagline
	d.LogoURL = in.LogoURL
	d.AccentColor = in.AccentColor
	if in.IsPublished != nil {
		d.IsPublished = *in.IsPublished
	}
	if in.IsIndexable != nil {
		d.IsIndexable = *in.IsIndexable
	}
	if err := ValidateDirectoryInput(d); err != nil {
		return badRequest(c, err)
	}
	if err := UpdateDirectory(d); err != nil {
		return serverError(c)
	}
	WriteAudit(ownerID, d.ID, "directory.update", "")
	return c.JSON(d)
}

func DeleteDirectoryHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	hard := c.Query("hard") == "true"
	if hard {
		if err := HardDeleteDirectory(id, ownerID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return notFound(c)
			}
			return serverError(c)
		}
		WriteAudit(ownerID, id, "directory.hard_delete", "")
		return c.SendStatus(fiber.StatusNoContent)
	}
	if err := SoftDeleteDirectory(id, ownerID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	WriteAudit(ownerID, id, "directory.soft_delete", "")
	return c.SendStatus(fiber.StatusNoContent)
}

type publishInput struct {
	IsPublished bool `json:"is_published"`
}

func PublishDirectoryHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	d, err := GetDirectoryForOwner(id, ownerID)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}
	var in publishInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	d.IsPublished = in.IsPublished
	if err := UpdateDirectory(d); err != nil {
		return serverError(c)
	}
	action := "directory.unpublish"
	if d.IsPublished {
		action = "directory.publish"
	}
	WriteAudit(ownerID, d.ID, action, "")
	return c.JSON(d)
}

type linkInput struct {
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	Value      string `json:"value"`
	Visibility string `json:"visibility"`
	RevealMode string `json:"reveal_mode"`
	Position   int    `json:"position"`
}

func AddLinkHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	if _, err := GetDirectoryForOwner(dirID, ownerID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	var in linkInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	l := &ContactLink{
		DirectoryID: dirID,
		Kind:        in.Kind,
		Label:       in.Label,
		Value:       in.Value,
		Visibility:  in.Visibility,
		RevealMode:  in.RevealMode,
		Position:    in.Position,
	}
	if err := ValidateContactLinkInput(l); err != nil {
		return badRequest(c, err)
	}
	if err := CreateLink(l); err != nil {
		return serverError(c)
	}
	WriteAudit(ownerID, dirID, "link.create", "kind="+l.Kind)
	return c.Status(fiber.StatusCreated).JSON(l)
}

func ListLinksHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	out, err := ListLinksForOwner(dirID, ownerID)
	if err != nil {
		return serverError(c)
	}
	return c.JSON(out)
}

func UpdateLinkHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	linkID, err := parseUint(c.Params("linkId"))
	if err != nil {
		return notFound(c)
	}
	l, err := GetLinkForOwner(linkID, dirID, ownerID)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}
	var in linkInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	if in.Kind != "" {
		l.Kind = in.Kind
	}
	if in.Label != "" {
		l.Label = in.Label
	}
	if in.Value != "" {
		l.Value = in.Value
	}
	if in.Visibility != "" {
		l.Visibility = in.Visibility
	}
	if in.RevealMode != "" {
		l.RevealMode = in.RevealMode
	}
	l.Position = in.Position
	if err := ValidateContactLinkInput(l); err != nil {
		return badRequest(c, err)
	}
	if err := UpdateLink(l); err != nil {
		return serverError(c)
	}
	WriteAudit(ownerID, dirID, "link.update", "id="+c.Params("linkId"))
	return c.JSON(l)
}

func DeleteLinkHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	linkID, err := parseUint(c.Params("linkId"))
	if err != nil {
		return notFound(c)
	}
	if err := DeleteLink(linkID, dirID, ownerID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	WriteAudit(ownerID, dirID, "link.delete", "id="+c.Params("linkId"))
	return c.SendStatus(fiber.StatusNoContent)
}

type reorderInput struct {
	Order []struct {
		ID       uint `json:"id"`
		Position int  `json:"position"`
	} `json:"order"`
}

func ReorderLinksHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	var in reorderInput
	if err := c.BodyParser(&in); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	positions := make(map[uint]int, len(in.Order))
	for _, o := range in.Order {
		positions[o.ID] = o.Position
	}
	if err := ReorderLinks(dirID, ownerID, positions); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	WriteAudit(ownerID, dirID, "link.reorder", "")
	return c.SendStatus(fiber.StatusNoContent)
}

func ListAuditHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirID, err := parseUint(c.Params("id"))
	if err != nil {
		return notFound(c)
	}
	out, err := ListAudit(dirID, ownerID, 100)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	return c.JSON(out)
}

// --- Public endpoints --------------------------------------------------------

func PublicGetHandler(c *fiber.Ctx) error {
	slug := strings.ToLower(c.Params("slug"))
	d, err := GetPublishedDirectoryBySlug(slug)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}

	applyPrivacyHeaders(c, d.IsIndexable)
	c.Set("Cache-Control", "public, max-age=60")

	if !respectsDNT(c) && shouldCountView(slug, c.IP()) {
		go func(s string) { _ = IncrementViews(s) }(slug)
	}

	view := PublicView{
		Slug:        d.Slug,
		Name:        d.Name,
		Tagline:     d.Tagline,
		LogoURL:     d.LogoURL,
		AccentColor: d.AccentColor,
	}
	for _, l := range d.Links {
		pv := PublicLinkView{
			ID:         l.ID,
			Kind:       l.Kind,
			Label:      l.Label,
			Position:   l.Position,
			RevealMode: l.RevealMode,
			Verified:   l.VerifiedAt != nil,
		}
		if !l.ShouldHideValue() {
			pv.Value = l.Value
		}
		view.Links = append(view.Links, pv)
	}
	return c.JSON(view)
}

func PublicLinkValueHandler(c *fiber.Ctx) error {
	slug := strings.ToLower(c.Params("slug"))
	linkID, err := parseUint(c.Params("linkId"))
	if err != nil {
		return notFound(c)
	}
	l, err := GetPublicLinkValue(slug, linkID)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}
	applyPrivacyHeaders(c, false)
	c.Set("Cache-Control", "private, no-store")
	return c.JSON(fiber.Map{
		"id":    l.ID,
		"kind":  l.Kind,
		"value": l.Value,
	})
}

func PublicQRHandler(c *fiber.Ctx) error {
	slug := strings.ToLower(c.Params("slug"))
	if _, err := GetPublishedDirectoryBySlug(slug); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFound(c)
		}
		return serverError(c)
	}
	size := 256
	if v := c.Query("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			size = n
		}
	}
	png, err := GenerateQR(PublicURLForSlug(slug), size)
	if err != nil {
		return serverError(c)
	}
	applyPrivacyHeaders(c, true)
	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "public, max-age=3600")
	return c.Send(png)
}

func PublicVCardHandler(c *fiber.Ctx) error {
	slug := strings.ToLower(c.Params("slug"))
	d, err := GetPublishedDirectoryBySlug(slug)
	if errors.Is(err, ErrNotFound) {
		return notFound(c)
	}
	if err != nil {
		return serverError(c)
	}
	applyPrivacyHeaders(c, false)
	c.Set("Content-Type", "text/vcard; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=\""+d.Slug+".vcf\"")
	c.Set("Cache-Control", "private, no-store")
	return c.SendString(RenderVCard(d))
}

// --- Account-scoped endpoints ------------------------------------------------

func ExportMeHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	dirs, err := ListDirectoriesByOwner(ownerID)
	if err != nil {
		return serverError(c)
	}
	out := make([]Directory, 0, len(dirs))
	for i := range dirs {
		full, err := GetDirectoryForOwner(dirs[i].ID, ownerID)
		if err != nil {
			continue
		}
		out = append(out, *full)
	}
	c.Set("Cache-Control", "private, no-store")
	return c.JSON(fiber.Map{"directories": out})
}

func DeleteMeHandler(c *fiber.Ctx) error {
	ownerID, ok := currentUserID(c)
	if !ok {
		return notFound(c)
	}
	if err := PurgeOwner(ownerID); err != nil {
		return serverError(c)
	}
	if err := auth.PurgeUser(ownerID); err != nil {
		return serverError(c)
	}
	auth.ClearSessionCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}
