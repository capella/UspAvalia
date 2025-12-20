package models

import "time"

type User struct {
	ID                      uint       `gorm:"primaryKey"                                     json:"id"`
	EmailHash               string     `gorm:"column:email_hash;uniqueIndex;not null;size:64" json:"-"`
	EmailVerified           bool       `gorm:"default:false"                                  json:"email_verified"`
	EmailVerificationToken  string     `gorm:"size:255"                                       json:"-"`
	EmailVerificationExpiry *time.Time `                                                      json:"-"`
	CreatedAt               time.Time  `                                                      json:"created_at"`
	UpdatedAt               time.Time  `                                                      json:"updated_at"`
}