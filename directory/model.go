package directory

import (
	"time"

	"gorm.io/gorm"
)

type Directory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	OwnerID     uint           `gorm:"index;not null" json:"-"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"not null" json:"name"`
	Tagline     string         `json:"tagline"`
	LogoURL     string         `json:"logo_url"`
	AccentColor string         `json:"accent_color"`
	IsPublished bool           `gorm:"not null;default:false" json:"is_published"`
	IsIndexable bool           `gorm:"not null;default:false" json:"is_indexable"`
	Views       uint64         `gorm:"not null;default:0" json:"views"`

	Links []ContactLink `gorm:"constraint:OnDelete:CASCADE" json:"links,omitempty"`
}

type ContactLink struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	DirectoryID uint           `gorm:"index;not null" json:"-"`
	Kind        string         `gorm:"not null" json:"kind"`
	Label       string         `json:"label"`
	Value       string         `gorm:"not null" json:"value,omitempty"`
	Visibility  string         `gorm:"not null;default:'public'" json:"visibility"`
	RevealMode  string         `gorm:"not null;default:'auto'" json:"reveal_mode"`
	Position    int            `gorm:"not null;default:0" json:"position"`
	VerifiedAt  *time.Time     `json:"verified_at,omitempty"`
}

type AuditEntry struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"at"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	DirectoryID uint      `gorm:"index" json:"directory_id"`
	Action      string    `gorm:"not null" json:"action"`
	Summary     string    `json:"summary"`
}

// PublicView is the serialized shape returned by /api/v1/c/:slug.
// Owner identity and timestamps are deliberately stripped.
type PublicView struct {
	Slug        string             `json:"slug"`
	Name        string             `json:"name"`
	Tagline     string             `json:"tagline,omitempty"`
	LogoURL     string             `json:"logo_url,omitempty"`
	AccentColor string             `json:"accent_color,omitempty"`
	Links       []PublicLinkView   `json:"links"`
}

type PublicLinkView struct {
	ID         uint   `json:"id"`
	Kind       string `json:"kind"`
	Label      string `json:"label,omitempty"`
	Position   int    `json:"position"`
	RevealMode string `json:"reveal_mode"`
	Value      string `json:"value,omitempty"`
	Verified   bool   `json:"verified"`
}
