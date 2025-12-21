package models

type VoteType int

const (
	VoteTypeGeneral         VoteType = 1
	VoteTypeTeaching        VoteType = 2
	VoteTypeCommitment      VoteType = 3
	VoteTypeStudentRelation VoteType = 4
	VoteTypeDifficulty      VoteType = 5
)

func (vt VoteType) String() string {
	switch vt {
	case VoteTypeGeneral:
		return "general"
	case VoteTypeTeaching:
		return "teaching"
	case VoteTypeCommitment:
		return "commitment"
	case VoteTypeStudentRelation:
		return "student_relation"
	case VoteTypeDifficulty:
		return "difficulty"
	default:
		return "unknown"
	}
}

type VoteTypeStats struct {
	Type  VoteType `json:"type"`
	Count int64    `json:"count"`
	Std   float64  `json:"std"`
	Avg   float64  `json:"avg"`
}

type Vote struct {
	ID               uint   `gorm:"primaryKey"                                                                            json:"id"`
	ClassProfessorID uint   `gorm:"not null"                                                                              json:"class_professor_id"`
	UserID           string `gorm:"size:255;not null"                                                                     json:"user_id"`
	Time             int64  `gorm:"not null"                                                                              json:"time"`
	Score            int    `gorm:"not null"                                                                              json:"score"`
	Type             int    `gorm:"default:1;comment:'1-general;2-teaching;3-commitment;4-student_relation;5-difficulty'" json:"type"`

	ClassProfessor ClassProfessor `gorm:"foreignKey:ClassProfessorID" json:"class_professor,omitempty"`
}
