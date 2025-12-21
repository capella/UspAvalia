package models

import "time"

type Comment struct {
	ID               uint   `gorm:"primaryKey"             json:"id"`
	UserID           string `gorm:"size:255;not null"      json:"user_id"`
	Content          string `gorm:"size:800;not null"      json:"content"`
	ClassProfessorID uint   `gorm:"not null"               json:"class_professor_id"`
	Time             int64  `gorm:"not null"               json:"time"`

	ClassProfessor ClassProfessor `gorm:"foreignKey:ClassProfessorID" json:"class_professor,omitempty"`
	CommentVotes   []CommentVote  `gorm:"foreignKey:CommentID"        json:"comment_votes,omitempty"`
	CreatedAt      time.Time      `                                   json:"created_at"`
	UpdatedAt      time.Time      `                                   json:"updated_at"`
}
