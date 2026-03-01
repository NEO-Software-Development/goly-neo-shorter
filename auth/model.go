package auth

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:'user'"`
}

type Session struct {
	gorm.Model
	Token    string `gorm:"unique;not null"`
	UserID   uint   `gorm:"not null"`
	User     User
	ExpiresAt time.Time `gorm:"not null"`
}
