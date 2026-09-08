package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// VoteActivityData represents vote count for a specific timestamp
type VoteActivityData struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

// handleVoteActivity returns vote activity data for the heatmap
func (s *Server) handleVoteActivity(w http.ResponseWriter, r *http.Request) {
	// Get vote activity for the last year
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	// Format the day as a "YYYY-MM-DD" string in SQL. With parseTime=True
	// MySQL's DATE() comes back as time.Time, which cannot be scanned into
	// a string and used to leave the response empty.
	var dateExpression string
	dialectName := s.db.Dialector.Name()
	if dialectName == "sqlite" {
		dateExpression = "strftime('%Y-%m-%d', time, 'unixepoch')"
	} else {
		// MySQL and other databases
		dateExpression = "DATE_FORMAT(FROM_UNIXTIME(time), '%Y-%m-%d')"
	}

	// Check if filtering by class_professor ID
	idParam := r.URL.Query().Get("id")
	var query string
	var args []interface{}

	if idParam != "" {
		// Parse and validate the ID
		classProfessorID, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
			return
		}

		// Query votes for specific class/professor grouped by date
		query = fmt.Sprintf(`
			SELECT
				%s as date,
				COUNT(*) as count
			FROM votes
			WHERE time >= ? AND class_professor_id = ?
			GROUP BY %s
			ORDER BY date ASC
		`, dateExpression, dateExpression)
		args = []interface{}{oneYearAgo.Unix(), classProfessorID}
	} else {
		// Query all votes grouped by date
		query = fmt.Sprintf(`
			SELECT
				%s as date,
				COUNT(*) as count
			FROM votes
			WHERE time >= ?
			GROUP BY %s
			ORDER BY date ASC
		`, dateExpression, dateExpression)
		args = []interface{}{oneYearAgo.Unix()}
	}

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build response data
	activityMap := make(map[string]int)
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			continue
		}
		activityMap[date] = count
	}

	// Convert to array format expected by cal-heatmap
	// Cal-heatmap expects timestamp in seconds and value.
	// Start with an empty (non-nil) slice so the JSON is "[]", not "null".
	activity := []map[string]interface{}{}
	for date, count := range activityMap {
		// Parse date string to time
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}

		activity = append(activity, map[string]interface{}{
			"date":  t.Unix(),
			"value": count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}
