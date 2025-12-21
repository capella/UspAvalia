package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"                                     json:"id"`
	EmailHash string    `gorm:"column:email_hash;uniqueIndex;not null;size:64" json:"-"`
	CreatedAt time.Time `                                                      json:"created_at"`
	UpdatedAt time.Time `                                                      json:"updated_at"`
}
