package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"uspavalia/internal/models"

	"github.com/gorilla/mux"
)

// GET /api/matrusp/disciplines - Return all disciplines with turmas
func (s *Server) handleMatruspDisciplines(w http.ResponseWriter, r *http.Request) {
	// Get the latest updated_at timestamp from class_offerings to use as ETag basis
	var lastUpdated time.Time
	s.db.Model(&models.ClassOffering{}).
		Select("MAX(updated_at)").
		Scan(&lastUpdated)

	// Generate ETag based on last update time
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(lastUpdated.Format(time.RFC3339Nano))))

	// Check If-None-Match header
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	var disciplines []models.Discipline

	s.db.Preload("Unit").
		Preload("ClassOfferings").
		Find(&disciplines)

	// Convert to MatrUSP format
	result := make([]map[string]interface{}, len(disciplines))
	for i, disc := range disciplines {
		result[i] = s.disciplineToMatruspFormat(&disc)
	}

	// Set ETag header
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GET /api/matrusp/discipline/:code - Return specific discipline
func (s *Server) handleMatruspDiscipline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	var discipline models.Discipline
	err := s.db.Preload("Unit").
		Preload("ClassOfferings").
		Where("code = ?", code).
		First(&discipline).Error

	if err != nil {
		http.Error(w, "Discipline not found", http.StatusNotFound)
		return
	}

	// Generate ETag based on discipline's updated_at time
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(discipline.UpdatedAt.Format(time.RFC3339Nano))))

	// Check If-None-Match header
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	result := s.disciplineToMatruspFormat(&discipline)

	// Set ETag header
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) disciplineToMatruspFormat(disc *models.Discipline) map[string]interface{} {
	// Parse turmas and filter out past offerings
	turmas := make([]map[string]interface{}, 0)
	now := time.Now()

	for _, offering := range disc.ClassOfferings {
		// Skip offerings that have already ended
		if offering.EndDate != "" {
			endDate, err := parseDate(offering.EndDate)
			if err == nil && endDate.Before(now) {
				continue // Skip past offerings
			}
		}

		var horario []interface{}
		var vagas map[string]interface{}

		json.Unmarshal([]byte(offering.Schedules), &horario)
		json.Unmarshal([]byte(offering.Vacancies), &vagas)

		turmas = append(turmas, map[string]interface{}{
			"codigo":         offering.Code,
			"codigo_teorica": offering.TheoreticalCode,
			"inicio":         offering.StartDate,
			"fim":            offering.EndDate,
			"tipo":           offering.Type,
			"observacoes":    offering.Notes,
			"horario":        horario,
			"vagas":          vagas,
		})
	}

	unitName := ""
	if disc.Unit.ID > 0 {
		unitName = disc.Unit.Name
	}

	return map[string]interface{}{
		"codigo":            disc.Code,
		"nome":              disc.Name,
		"unidade":           unitName,
		"departamento":      disc.Department,
		"campus":            disc.Campus,
		"objetivos":         disc.Objectives,
		"programa_resumido": disc.Summary,
		"creditos_aula":     disc.CreditsClass,
		"creditos_trabalho": disc.CreditsWork,
		"turmas":            turmas,
	}
}

// parseDate parses a date in DD/MM/YYYY format
func parseDate(dateStr string) (time.Time, error) {
	parts := strings.Split(dateStr, "/")
	if len(parts) != 3 {
		return time.Time{}, nil
	}

	// Parse DD/MM/YYYY format
	layout := "02/01/2006"
	return time.Parse(layout, dateStr)
}
