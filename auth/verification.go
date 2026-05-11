package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"goly-app/database"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

const verificationTTL = 30 * time.Minute

// IssueVerificationToken creates a single-use token for the given user/kind
// and returns the plaintext to the caller. The plaintext should be delivered
// out-of-band (email, SMS) and is never persisted server-side.
func IssueVerificationToken(userID uint, kind string, targetID uint) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	entry := &VerificationToken{
		UserID:    userID,
		Kind:      kind,
		TargetID:  targetID,
		Hash:      hashVerificationToken(token),
		ExpiresAt: time.Now().Add(verificationTTL),
	}
	if err := database.DB.Create(entry).Error; err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeVerificationToken validates a token against the stored hash, marks
// it consumed, and returns the matched record. Returns ErrInvalidToken for
// mismatch, ErrExpiredToken for stale ones.
func ConsumeVerificationToken(token, kind string) (*VerificationToken, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	target := hashVerificationToken(token)
	var entry VerificationToken
	tx := database.DB.Where("hash = ? AND kind = ? AND consumed_at IS NULL", target, kind).First(&entry)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	if time.Now().After(entry.ExpiresAt) {
		return nil, ErrExpiredToken
	}
	now := time.Now()
	entry.ConsumedAt = &now
	if err := database.DB.Save(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func hashVerificationToken(token string) string {
	sum := sha256.Sum256([]byte("goly:vt:" + token))
	return hex.EncodeToString(sum[:])
}
