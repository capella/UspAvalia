package cmd

import (
	"uspavalia/internal/models"
	"uspavalia/internal/services"

	"gorm.io/gorm"
)

// Thin wrappers over the shared catalog helpers in internal/services, kept so
// the scraper commands read the same as before. See services/catalog.go for
// the dedupe rules.

func getOrCreateUnit(db *gorm.DB, name string) (models.Unit, error) {
	return services.GetOrCreateUnit(db, name)
}

func getOrCreateProfessor(db *gorm.DB, name string, unitID uint) (models.Professor, bool, error) {
	return services.GetOrCreateProfessor(db, name, unitID)
}

func getOrCreateDiscipline(db *gorm.DB, discipline models.Discipline) (models.Discipline, error) {
	return services.GetOrCreateDiscipline(db, discipline)
}

func getOrCreateClassProfessor(db *gorm.DB, classID, professorID uint) (models.ClassProfessor, bool, error) {
	return services.GetOrCreateClassProfessor(db, classID, professorID)
}
