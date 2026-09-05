package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"uspavalia/internal/models"

	"github.com/gorilla/mux"
)

const lookupLimit = 15

// LookupResult is one typeahead suggestion. Only the ID is ever sent back to
// the server: the label exists so the user can recognise what they picked.
type LookupResult struct {
	ID     uint   `json:"id"`
	Label  string `json:"label"`
	UnitID uint   `json:"unit_id,omitempty"`
}

// handleLookup serves GET /api/lookup/{kind}?q=...&unit_id=... for the
// "add discipline/professor" page. kind is units, disciplines or professors.
// It only returns rows that already exist; the page uses it so users select
// catalog entries instead of typing free text.
func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	kind := mux.Vars(r)["kind"]
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	unitID, _ := strconv.ParseUint(r.URL.Query().Get("unit_id"), 10, 32)

	pattern := "%" + escapeLike(query) + "%"
	results := []LookupResult{}

	switch kind {
	case "units":
		var units []models.Unit
		s.db.Where("name LIKE ? ESCAPE '!'", pattern).Order("name").Limit(lookupLimit).Find(&units)
		for _, u := range units {
			results = append(results, LookupResult{ID: u.ID, Label: u.Name})
		}
	case "disciplines":
		var disciplines []models.Discipline
		q := s.db.Where("name LIKE ? ESCAPE '!' OR code LIKE ? ESCAPE '!'", pattern, pattern)
		if unitID > 0 {
			q = q.Where("unit_id = ?", unitID)
		}
		q.Order("code").Limit(lookupLimit).Find(&disciplines)
		for _, d := range disciplines {
			results = append(results, LookupResult{
				ID:     d.ID,
				Label:  d.Code + " - " + d.Name,
				UnitID: d.UnitID,
			})
		}
	case "professors":
		var professors []models.Professor
		q := s.db.Where("name LIKE ? ESCAPE '!'", pattern)
		if unitID > 0 {
			q = q.Where("unit_id = ?", unitID)
		}
		q.Order("name").Limit(lookupLimit).Find(&professors)
		for _, p := range professors {
			results = append(results, LookupResult{ID: p.ID, Label: p.Name, UnitID: p.UnitID})
		}
	default:
		http.Error(w, "unknown lookup kind", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// likeEscape is the ESCAPE character for the LIKE patterns above. "!" is used
// instead of a backslash because SQLite and MySQL disagree on how a backslash
// inside a string literal is parsed.
const likeEscape = "!"

// escapeLike neutralises LIKE wildcards in user input so "%" or "_" typed by
// the user match literally. Pair it with `LIKE ? ESCAPE '!'`.
func escapeLike(s string) string {
	r := strings.NewReplacer(
		likeEscape, likeEscape+likeEscape,
		`%`, likeEscape+`%`,
		`_`, likeEscape+`_`,
	)
	return r.Replace(s)
}
