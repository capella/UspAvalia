package models

type CommentVote struct {
	ID        uint   `gorm:"primaryKey"                                       json:"id"`
	CommentID uint   `gorm:"not null;uniqueIndex:idx_comment_user"            json:"comment_id"`
	Time      int64  `gorm:"not null"                                         json:"time"`
	Vote      int    `gorm:"not null;comment:'-1 or 1'"                       json:"vote"`
	UserID    string `gorm:"size:255;not null;uniqueIndex:idx_comment_user"   json:"user_id"`

	Comment Comment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
}
