package handlers

import (
	"encoding/json"
	"net/http"
	"uspavalia/internal/models"

	"github.com/gorilla/csrf"
)

func (s *Server) handleTypeahead(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return JSON for typeahead
	w.Header().Set("Content-Type", "application/json")

	var disciplines []models.Discipline
	var professors []models.Professor

	searchPattern := "%" + query + "%"
	s.db.Preload("Unit").
		Where("name LIKE ? OR code LIKE ?", searchPattern, searchPattern).
		Limit(10).
		Find(&disciplines)
	s.db.Preload("Unit").Where("name LIKE ?", searchPattern).Limit(10).Find(&professors)

	var results []map[string]interface{}

	// Add disciplines (type 1)
	for _, discipline := range disciplines {
		results = append(results, map[string]interface{}{
			"id":   discipline.ID,
			"name": discipline.Code + " - " + discipline.Name,
			"type": 1,
		})
	}

	// Add professors (type 0 - matching original PHP)
	for _, professor := range professors {
		results = append(results, map[string]interface{}{
			"id":   professor.ID,
			"name": professor.Name,
			"type": 0,
		})
	}

	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("pesquisa")

	if query == "" {
		data := PageData{
			CSRFToken: csrf.Token(r),
			User:      s.getCurrentUser(r),
		}
		s.renderTemplate(w, r, "search", data)
		return
	}

	var disciplines []models.Discipline
	var professors []models.Professor

	searchPattern := "%" + query + "%"
	s.db.Preload("Unit").
		Where("name LIKE ? OR code LIKE ?", searchPattern, searchPattern).
		Limit(40).
		Find(&disciplines)
	s.db.Preload("Unit").Where("name LIKE ?", searchPattern).Limit(40).Find(&professors)

	// Calculate total result count
	resultCount := len(disciplines) + len(professors)

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"Query":       query,
			"Disciplines": disciplines,
			"Professors":  professors,
			"ResultCount": resultCount,
		},
	}

	s.renderTemplate(w, r, "search", data)
}