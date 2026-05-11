package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"goly-app/database"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

// dummyHash is used to flatten timing differences between "user not found"
// and "wrong password": we always run an Argon2 comparison even when the
// username is unknown. The hash itself decodes to a valid argon2id record
// for a value no client will guess.
var dummyHash = ""

func init() {
	h, err := HashPassword("a-very-long-static-value-no-user-can-guess-purposefully")
	if err == nil {
		dummyHash = h
	}
}

func Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse request",
		})
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username, password, and email are required",
		})
	}
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters",
		})
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email must be a valid address",
		})
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	user := &User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Email:        req.Email,
	}

	if err := CreateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "user created successfully",
	})
}

// Login uses constant-time, ambiguous error responses: unknown-username and
// wrong-password both return the same 401 with the same body, and the
// password hashing step runs in both branches so timing stays similar.
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse request",
		})
	}

	user, err := GetUserByUsername(req.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_, _ = ComparePasswordAndHash(req.Password, dummyHash)
		return ambiguousAuthFailure(c)
	}
	if err != nil {
		_, _ = ComparePasswordAndHash(req.Password, dummyHash)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed",
		})
	}

	match, _ := ComparePasswordAndHash(req.Password, user.PasswordHash)
	if !match {
		return ambiguousAuthFailure(c)
	}

	if err := CheckTOTPOrBackup(user, req.OTP); err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":         "otp_required_or_invalid",
				"totp_required": true,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed",
		})
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate session token",
		})
	}

	device := deviceHash(c)
	session := &Session{
		Token:      token,
		UserID:     user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		DeviceHash: device,
		Label:      truncate(c.Get("User-Agent"), 120),
	}

	if err := CreateSession(session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	if !hasSeenDevice(user.ID, device) {
		// Coarse new-device signal — UA fingerprint without IP. In production
		// this would emit an email; for now we record it in the server log so
		// the owner's audit + session list can show "new device" badges.
		log.Printf("auth: new-device login for user_id=%d", user.ID)
	}

	SetSessionCookie(c, token, session.ExpiresAt)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "login successful",
	})
}

func ambiguousAuthFailure(c *fiber.Ctx) error {
	// Burn a small amount of CPU regardless of branch so login latency stays
	// in the same ballpark for failure paths.
	_ = argon2.IDKey([]byte("x"), []byte("xxxxxxxxxxxxxxxx"), 1, 16*1024, 1, 16)
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "invalid_credentials",
	})
}

func Logout(c *fiber.Ctx) error {
	token := c.Cookies("session_token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := DeleteSessionByToken(token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to logout",
		})
	}

	ClearSessionCookie(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logout successful",
	})
}

// SetSessionCookie centralizes the session-cookie attributes. Secure/HttpOnly
// and SameSite=Lax are applied uniformly so the login and logout paths can't
// drift apart.
func SetSessionCookie(c *fiber.Ctx, token string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})
}

func ClearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})
}

// --- 2FA endpoints -----------------------------------------------------------

type otpVerifyRequest struct {
	OTP      string `json:"otp"`
	Password string `json:"password"`
}

func Enroll2FA(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	url, err := EnrollTOTP(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	return c.JSON(fiber.Map{"otpauth_url": url})
}

func Verify2FA(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req otpVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	codes, err := VerifyTOTP(user, req.OTP)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_otp"})
	}
	return c.JSON(fiber.Map{
		"enabled":      true,
		"backup_codes": codes,
	})
}

// deviceHash returns a coarse fingerprint of the calling device so the server
// can distinguish "logged in from same browser" vs "new device". Only the
// User-Agent is used; IPs are deliberately excluded so visitor identity is
// never persisted server-side. The hash is keyed with PASSWORD_PEPPER (if set)
// so the values are useless without server access.
func deviceHash(c *fiber.Ctx) string {
	ua := c.Get("User-Agent")
	pepper := os.Getenv("PASSWORD_PEPPER")
	h := sha256.Sum256([]byte("goly:dev:" + pepper + "|" + ua))
	return hex.EncodeToString(h[:16])
}

func hasSeenDevice(userID uint, device string) bool {
	var count int64
	if err := database.DB.Model(&Session{}).Where("user_id = ? AND device_hash = ?", userID, device).Count(&count).Error; err == nil {
		return count > 1
	}
	return false
}

func parseUintParam(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// --- Email verification ------------------------------------------------------

type startEmailRequest struct {
	Email string `json:"email"`
}

type confirmEmailRequest struct {
	Token string `json:"token"`
}

func StartEmailVerification(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req startEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email != "" {
		user.Email = req.Email
		user.EmailVerifiedAt = nil
		if err := UpdateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
		}
	}
	if user.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email required"})
	}
	token, err := IssueVerificationToken(user.ID, "email", 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	resp := fiber.Map{"sent": true}
	// In dev (no SMTP wired) we expose the token inline so the flow is testable.
	if os.Getenv("EMAIL_PROVIDER") == "" {
		resp["dev_token"] = token
	}
	return c.JSON(resp)
}

func ConfirmEmailVerification(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req confirmEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	entry, err := ConsumeVerificationToken(req.Token, "email")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if entry.UserID != user.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token mismatch"})
	}
	now := time.Now()
	user.EmailVerifiedAt = &now
	if err := UpdateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	return c.JSON(fiber.Map{"verified": true})
}

// --- Session list / revoke ---------------------------------------------------

type sessionView struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Label      string    `json:"label"`
	DeviceHash string    `json:"device_hash"`
	Current    bool      `json:"current"`
}

func ListSessions(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	current := c.Cookies("session_token")
	sessions, err := ListSessionsForUser(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	out := make([]sessionView, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionView{
			ID:         s.ID,
			CreatedAt:  s.CreatedAt,
			ExpiresAt:  s.ExpiresAt,
			Label:      s.Label,
			DeviceHash: s.DeviceHash,
			Current:    s.Token == current,
		})
	}
	return c.JSON(out)
}

func RevokeSession(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := parseUintParam(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad id"})
	}
	if err := DeleteSessionForUser(id, user.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func Disable2FA(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req otpVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	// Re-authenticate with password before disabling 2FA.
	match, _ := ComparePasswordAndHash(req.Password, user.PasswordHash)
	if !match {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_credentials"})
	}
	if err := CheckTOTPOrBackup(user, req.OTP); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_otp"})
	}
	if err := DisableTOTP(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
	}
	return c.JSON(fiber.Map{"enabled": false})
}
