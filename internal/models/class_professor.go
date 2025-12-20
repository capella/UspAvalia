package models

import (
	"math"
	"time"

	"gorm.io/gorm"
)

type ClassProfessor struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ClassID     uint   `gorm:"not null"   json:"class_id"`
	ProfessorID uint   `gorm:"not null"   json:"professor_id"`
	Usage       string `gorm:"size:200"   json:"usage,omitempty"`
	Time        *int64 `                  json:"time,omitempty"`

	Discipline Discipline `gorm:"foreignKey:ClassID"          json:"discipline,omitempty"`
	Professor  Professor  `gorm:"foreignKey:ProfessorID"      json:"professor,omitempty"`
	Votes      []Vote     `gorm:"foreignKey:ClassProfessorID" json:"votes,omitempty"`
	Comments   []Comment  `gorm:"foreignKey:ClassProfessorID" json:"comments,omitempty"`
	CreatedAt  time.Time  `                                   json:"created_at"`
	UpdatedAt  time.Time  `                                   json:"updated_at"`
}


func (cp *ClassProfessor) CalculateStatsByType(db *gorm.DB) ([]VoteTypeStats, error) {
	var results []struct {
		Type      int     `gorm:"column:type"`
		Count     int64   `gorm:"column:count"`
		Avg       float64 `gorm:"column:avg"`
		AvgSquare float64 `gorm:"column:avg_square"`
	}

	err := db.Model(&Vote{}).
		Select("type, COUNT(*) as count, AVG(score)*2 as avg, AVG(score * score) as avg_square").
		Where("class_professor_id = ?", cp.ID).
		Group("class_professor_id, type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make([]VoteTypeStats, len(results))
	for i, result := range results {
		// Calculate standard deviation: σ = √(E[X²] - E[X]²)
		avgScore := result.Avg / 2 // Convert back to original scale for variance calculation
		variance := result.AvgSquare - (avgScore * avgScore)
		std := 0.0
		if variance > 0 {
			std = math.Sqrt(variance) * 2 // Apply the *2 multiplier
		}

		stats[i] = VoteTypeStats{
			Type:  VoteType(result.Type),
			Count: result.Count,
			Std:   math.Round(std*100) / 100,
			Avg:   math.Round(result.Avg*100) / 100,
		}
	}

	return stats, nil
}

