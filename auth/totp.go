package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"goly-app/database"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

const backupCodeLen = 10
const backupCodeCount = 10

var ErrTOTPNotEnabled = errors.New("totp not enabled")
var ErrInvalidOTP = errors.New("invalid otp")

// EnrollTOTP generates a fresh secret + provisioning URL for the user. The
// secret is stored on the user but TOTPEnabledAt is only set after a
// successful VerifyTOTP. Returns the otpauth:// URL the client renders as a
// QR for the authenticator app.
func EnrollTOTP(user *User) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Goly",
		AccountName: user.Username,
	})
	if err != nil {
		return "", err
	}
	user.TOTPSecret = key.Secret()
	user.TOTPEnabledAt = nil
	if err := database.DB.Save(user).Error; err != nil {
		return "", err
	}
	return key.URL(), nil
}

// VerifyTOTP confirms a freshly enrolled secret and turns 2FA on.
func VerifyTOTP(user *User, code string) ([]string, error) {
	if user.TOTPSecret == "" {
		return nil, ErrTOTPNotEnabled
	}
	if !totp.Validate(code, user.TOTPSecret) {
		return nil, ErrInvalidOTP
	}
	now := time.Now()
	user.TOTPEnabledAt = &now
	if err := database.DB.Save(user).Error; err != nil {
		return nil, err
	}
	return regenerateBackupCodes(user.ID)
}

// CheckTOTPOrBackup validates a login-time OTP, accepting either a current
// TOTP value or a single-use backup code.
func CheckTOTPOrBackup(user *User, code string) error {
	if user.TOTPSecret == "" || user.TOTPEnabledAt == nil {
		return nil
	}
	if code == "" {
		return ErrInvalidOTP
	}
	if totp.Validate(code, user.TOTPSecret) {
		return nil
	}
	return consumeBackupCode(user.ID, code)
}

// DisableTOTP wipes the user's secret and revokes all backup codes.
func DisableTOTP(user *User) error {
	user.TOTPSecret = ""
	user.TOTPEnabledAt = nil
	if err := database.DB.Save(user).Error; err != nil {
		return err
	}
	return database.DB.Where("user_id = ?", user.ID).Delete(&BackupCode{}).Error
}

func regenerateBackupCodes(userID uint) ([]string, error) {
	if err := database.DB.Where("user_id = ?", userID).Delete(&BackupCode{}).Error; err != nil {
		return nil, err
	}
	plain := make([]string, 0, backupCodeCount)
	for i := 0; i < backupCodeCount; i++ {
		code, err := newBackupCode()
		if err != nil {
			return nil, err
		}
		plain = append(plain, code)
		entry := &BackupCode{UserID: userID, Hash: hashBackupCode(code)}
		if err := database.DB.Create(entry).Error; err != nil {
			return nil, err
		}
	}
	return plain, nil
}

func newBackupCode() (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	enc := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw))
	if len(enc) > backupCodeLen {
		enc = enc[:backupCodeLen]
	}
	return enc[:5] + "-" + enc[5:], nil
}

func hashBackupCode(code string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	sum := sha256.Sum256([]byte("goly:bc:" + normalized))
	return hex.EncodeToString(sum[:])
}

func consumeBackupCode(userID uint, code string) error {
	target := hashBackupCode(code)
	var codes []BackupCode
	if err := database.DB.Where("user_id = ? AND used_at IS NULL", userID).Find(&codes).Error; err != nil {
		return err
	}
	// Constant-time scan: walk every stored code so a matching one is found
	// in the same time as a non-match.
	var match *BackupCode
	for i := range codes {
		if subtle.ConstantTimeCompare([]byte(codes[i].Hash), []byte(target)) == 1 {
			match = &codes[i]
		}
	}
	if match == nil {
		return ErrInvalidOTP
	}
	now := time.Now()
	match.UsedAt = &now
	return database.DB.Save(match).Error
}
