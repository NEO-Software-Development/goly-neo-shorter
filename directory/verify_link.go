package directory

import (
	"errors"
	"goly-app/auth"
	"goly-app/database"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// StartLinkVerificationHandler issues a verification token addressed to the
// value stored on a contact link (the email/phone the owner advertised on
// their public page). Delivering the token to that channel proves the
// channel is theirs.
//
// Send-side delivery is intentionally pluggable: when EMAIL_PROVIDER /
// SMS_PROVIDER are unset, the token is echoed in the response so flows can
// be tested locally without a paid provider.
func StartLinkVerificationHandler(c *fiber.Ctx) error {
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
	if !canVerifyKind(l.Kind) {
		return badRequest(c, errors.New("this contact kind cannot be verified yet"))
	}
	token, err := auth.IssueVerificationToken(ownerID, "contact", l.ID)
	if err != nil {
		return serverError(c)
	}
	resp := fiber.Map{
		"sent":      true,
		"channel":   l.Kind,
		"target":    l.Value,
		"expires_in": int(time.Minute * 30 / time.Second),
	}
	if (l.Kind == "email" && os.Getenv("EMAIL_PROVIDER") == "") ||
		((l.Kind == "phone" || l.Kind == "sms" || l.Kind == "whatsapp") && os.Getenv("SMS_PROVIDER") == "") {
		resp["dev_token"] = token
	}
	return c.JSON(resp)
}

type confirmLinkRequest struct {
	Token string `json:"token"`
}

func ConfirmLinkVerificationHandler(c *fiber.Ctx) error {
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
	var req confirmLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, errors.New("invalid JSON"))
	}
	entry, err := auth.ConsumeVerificationToken(req.Token, "contact")
	if err != nil {
		return badRequest(c, err)
	}
	if entry.UserID != ownerID || entry.TargetID != l.ID {
		return badRequest(c, errors.New("token mismatch"))
	}
	now := time.Now()
	l.VerifiedAt = &now
	if err := database.DB.Save(l).Error; err != nil {
		return serverError(c)
	}
	WriteAudit(ownerID, dirID, "link.verify", "id="+c.Params("linkId"))
	return c.JSON(l)
}

func canVerifyKind(k string) bool {
	switch k {
	case "email", "phone", "sms", "whatsapp":
		return true
	}
	return false
}
