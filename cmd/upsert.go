package cmd

import (
	"errors"
	"fmt"
	"strings"
	"uspavalia/internal/models"

	"gorm.io/gorm"
)

// The helpers below are the single find-or-create path for scraper commands.
// They trim names and only create on gorm.ErrRecordNotFound: any other error
// (e.g. a bad column name) is returned instead of silently creating a
// duplicate row, which is how the units table once grew to 6,430 rows for 80
// real units.

func getOrCreateUnit(db *gorm.DB, name string) (models.Unit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Unit{}, fmt.Errorf("empty unit name")
	}

	var unit models.Unit
	err := db.Where("name = ?", name).First(&unit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		unit = models.Unit{Name: name}
		err = db.Create(&unit).Error
	}
	return unit, err
}

func getOrCreateProfessor(db *gorm.DB, name string, unitID uint) (models.Professor, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Professor{}, false, fmt.Errorf("empty professor name")
	}

	var professor models.Professor
	err := db.Where("name = ?", name).First(&professor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		professor = models.Professor{Name: name, UnitID: unitID}
		return professor, true, db.Create(&professor).Error
	}
	return professor, false, err
}

// getOrCreateDiscipline finds a discipline by code (its identity) and updates
// its fields, or creates it.
func getOrCreateDiscipline(db *gorm.DB, discipline models.Discipline) (models.Discipline, error) {
	discipline.Code = strings.TrimSpace(discipline.Code)
	discipline.Name = strings.TrimSpace(discipline.Name)
	if discipline.Code == "" {
		return discipline, fmt.Errorf("empty discipline code")
	}

	var existing models.Discipline
	err := db.Where("code = ?", discipline.Code).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return discipline, db.Create(&discipline).Error
	}
	if err != nil {
		return discipline, err
	}
	err = db.Model(&existing).Updates(discipline).Error
	discipline.ID = existing.ID
	return discipline, err
}

func getOrCreateClassProfessor(db *gorm.DB, classID, professorID uint) (models.ClassProfessor, bool, error) {
	var cp models.ClassProfessor
	err := db.Where("class_id = ? AND professor_id = ?", classID, professorID).First(&cp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cp = models.ClassProfessor{ClassID: classID, ProfessorID: professorID}
		return cp, true, db.Create(&cp).Error
	}
	return cp, false, err
}
