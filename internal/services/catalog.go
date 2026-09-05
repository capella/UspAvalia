package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"uspavalia/internal/models"

	"gorm.io/gorm"
)

// The helpers below are the single find-or-create path for the catalog
// (units, disciplines, professors, class-professor pairs and offerings). They
// are shared by the scraper commands and the "add discipline/professor" page,
// so both go through the same dedupe rules. They trim names and only create on
// gorm.ErrRecordNotFound: any other error (e.g. a bad column name) is returned
// instead of silently creating a duplicate row, which is how the units table
// once grew to 6,430 rows for 80 real units.

func GetOrCreateUnit(db *gorm.DB, name string) (models.Unit, error) {
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

func GetOrCreateProfessor(db *gorm.DB, name string, unitID uint) (models.Professor, bool, error) {
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

// GetOrCreateDiscipline finds a discipline by code (its identity) and updates
// its fields, or creates it.
func GetOrCreateDiscipline(db *gorm.DB, discipline models.Discipline) (models.Discipline, error) {
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

func GetOrCreateClassProfessor(db *gorm.DB, classID, professorID uint) (models.ClassProfessor, bool, error) {
	var cp models.ClassProfessor
	err := db.Where("class_id = ? AND professor_id = ?", classID, professorID).First(&cp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cp = models.ClassProfessor{ClassID: classID, ProfessorID: professorID}
		return cp, true, db.Create(&cp).Error
	}
	return cp, false, err
}

// Semester is a USP half-year: Year plus Half (1 = Jan–Jun, 2 = Jul–Dec).
type Semester struct {
	Year int
	Half int
}

// UserOfferingSuffix closes the turma code of offerings created from the
// website. Jupiter numbers turmas from 01, so YYYYH00 never collides with a
// scraped turma and still groups with the real ones of the same semester.
const UserOfferingSuffix = "00"

// UserOfferingNote is stored in ClassOffering.Notes for website-created rows.
const UserOfferingNote = "Turma adicionada por usuário do USP Avalia"

// ScheduleEntry mirrors the scraper's HorarioInfo so offerings created from
// the website have the exact JSON shape the scraper and MatrUSP expect.
type ScheduleEntry struct {
	Day        string   `json:"dia"`
	Start      string   `json:"inicio"`
	End        string   `json:"fim"`
	Professors []string `json:"professores"`
}

// CurrentSemester returns the semester containing t.
func CurrentSemester(t time.Time) Semester {
	half := 1
	if t.Month() >= time.July {
		half = 2
	}
	return Semester{Year: t.Year(), Half: half}
}

// Valid reports whether the semester is a plausible USP half-year.
func (s Semester) Valid(now time.Time) bool {
	return s.Year >= 2000 && s.Year <= now.Year()+1 && (s.Half == 1 || s.Half == 2)
}

// Prefix is the five-character prefix Jupiter uses on turma codes: YYYYH.
func (s Semester) Prefix() string {
	return fmt.Sprintf("%d%d", s.Year, s.Half)
}

// Code is the turma code used for offerings created from the website.
func (s Semester) Code() string {
	return s.Prefix() + UserOfferingSuffix
}

func (s Semester) String() string {
	return fmt.Sprintf("%d/%d", s.Year, s.Half)
}

// ParseSemesterCode reads the semester out of a Jupiter turma code (YYYYH...).
func ParseSemesterCode(code string) (Semester, bool) {
	if len(code) < 5 {
		return Semester{}, false
	}
	var sem Semester
	if _, err := fmt.Sscanf(code[:5], "%4d%1d", &sem.Year, &sem.Half); err != nil {
		return Semester{}, false
	}
	if sem.Half != 1 && sem.Half != 2 {
		return Semester{}, false
	}
	return sem, true
}

// fallbackDates approximates the USP calendar when no scraped offering of the
// semester exists to copy the real dates from.
func (s Semester) fallbackDates() (string, string) {
	if s.Half == 1 {
		return fmt.Sprintf("01/02/%d", s.Year), fmt.Sprintf("15/07/%d", s.Year)
	}
	return fmt.Sprintf("01/08/%d", s.Year), fmt.Sprintf("20/12/%d", s.Year)
}

// SemesterDates returns the start and end dates (DD/MM/YYYY) for the semester,
// taken from the most common dates among scraped offerings of that semester,
// falling back to an approximate calendar.
func SemesterDates(db *gorm.DB, sem Semester) (string, string) {
	var row struct {
		StartDate string
		EndDate   string
	}
	err := db.Model(&models.ClassOffering{}).
		Select("start_date, end_date").
		Where("code LIKE ? AND code NOT LIKE ? AND start_date <> '' AND end_date <> ''",
			sem.Prefix()+"%", sem.Prefix()+UserOfferingSuffix).
		Group("start_date, end_date").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil || row.StartDate == "" || row.EndDate == "" {
		return sem.fallbackDates()
	}
	return row.StartDate, row.EndDate
}

// EnsureUserOffering records that professor taught discipline in the given
// semester by creating (or extending) the website-created ClassOffering for
// that discipline and semester. The professor is listed in the schedule JSON
// the same way the scraper does, so MatrUSP and future scrapes treat the row
// like any other turma. Returns the offering and whether it was created.
func EnsureUserOffering(
	db *gorm.DB,
	discipline models.Discipline,
	professor models.Professor,
	sem Semester,
) (models.ClassOffering, bool, error) {
	var offering models.ClassOffering
	err := db.Where("discipline_id = ? AND code = ?", discipline.ID, sem.Code()).
		First(&offering).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		start, end := SemesterDates(db, sem)
		schedules, _ := json.Marshal([]ScheduleEntry{{Professors: []string{professor.Name}}})
		offering = models.ClassOffering{
			DisciplineID: discipline.ID,
			Code:         sem.Code(),
			StartDate:    start,
			EndDate:      end,
			Type:         "Teórica",
			Notes:        UserOfferingNote,
			Schedules:    string(schedules),
			Vacancies:    "{}",
		}
		return offering, true, db.Create(&offering).Error
	}
	if err != nil {
		return offering, false, err
	}

	var entries []ScheduleEntry
	if offering.Schedules != "" {
		if err := json.Unmarshal([]byte(offering.Schedules), &entries); err != nil {
			entries = nil
		}
	}
	for _, entry := range entries {
		for _, name := range entry.Professors {
			if name == professor.Name {
				return offering, false, nil
			}
		}
	}
	if len(entries) == 0 {
		entries = []ScheduleEntry{{}}
	}
	entries[0].Professors = append(entries[0].Professors, professor.Name)
	schedules, _ := json.Marshal(entries)
	offering.Schedules = string(schedules)
	return offering, false, db.Model(&offering).Update("schedules", offering.Schedules).Error
}

// ProfessorSemesters lists, newest first, the semesters in which professor
// appears in the discipline's offerings.
func ProfessorSemesters(db *gorm.DB, disciplineID uint, professorName string) []Semester {
	var offerings []models.ClassOffering
	db.Select("code, schedules").
		Where("discipline_id = ?", disciplineID).
		Order("code DESC").
		Find(&offerings)

	seen := map[Semester]bool{}
	var result []Semester
	for _, offering := range offerings {
		sem, ok := ParseSemesterCode(offering.Code)
		if !ok || seen[sem] {
			continue
		}
		var entries []ScheduleEntry
		if json.Unmarshal([]byte(offering.Schedules), &entries) != nil {
			continue
		}
		for _, entry := range entries {
			for _, name := range entry.Professors {
				if name == professorName {
					seen[sem] = true
					result = append(result, sem)
				}
			}
		}
	}
	return result
}
