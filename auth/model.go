package auth

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username        string `gorm:"unique;not null"`
	PasswordHash    string `gorm:"not null"`
	Role            string `gorm:"not null;default:'user'"`
	Email           string `gorm:"index"`
	EmailVerifiedAt *time.Time
	TOTPSecret      string
	TOTPEnabledAt   *time.Time
}

type Session struct {
	gorm.Model
	Token      string `gorm:"unique;not null"`
	UserID     uint   `gorm:"not null"`
	User       User
	ExpiresAt  time.Time `gorm:"not null"`
	DeviceHash string    `gorm:"index"`
	Label      string
}

// VerificationToken holds short-lived, hashed challenge tokens for email and
// contact-channel verification. The plaintext is never stored; we hash with
// SHA-256 so lookups are constant cost and a DB leak can't be replayed.
type VerificationToken struct {
	gorm.Model
	UserID    uint   `gorm:"index;not null"`
	Kind      string `gorm:"not null"` // "email" | "contact"
	TargetID  uint   // optional: contact_link.id for "contact"
	Hash      string `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	ConsumedAt *time.Time
}

// BackupCode stores a hashed one-time recovery code. The plaintext is shown to
// the user exactly once when 2FA is enabled and is never persisted.
type BackupCode struct {
	gorm.Model
	UserID uint   `gorm:"index;not null"`
	Hash   string `gorm:"unique;not null"`
	UsedAt *time.Time
}
