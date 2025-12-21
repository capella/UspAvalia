package models

import "time"

// LoginToken represents a one-time use magic link login token
type LoginToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null;size:255" json:"-"`
	EmailHash string    `gorm:"index;not null;size:64" json:"-"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
