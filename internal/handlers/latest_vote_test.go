package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uspavalia/internal/models"
)

// TestStatsAndPrefillUseLatestVote covers re-voting: the professor page
// counts the user once with their newest score, and the rating modal is
// preselected with that score so they can change it.
func TestStatsAndPrefillUseLatestVote(t *testing.T) {
	s := newAuthServer(t)
	cp := seedClassProfessor(t, s.db)
	s.db.Create(&models.User{EmailHash: "u"})

	now := time.Now().Unix()
	votes := []models.Vote{
		{ClassProfessorID: cp.ID, UserID: "1", Type: 5, Score: 2, Time: now - 60},
		{ClassProfessorID: cp.ID, UserID: "1", Type: 5, Score: 4, Time: now},
		{ClassProfessorID: cp.ID, UserID: "1", Type: 1, Score: 3, Time: now},
	}
	if err := s.db.Create(&votes).Error; err != nil {
		t.Fatalf("create votes: %v", err)
	}

	stats, err := cp.CalculateStatsByType(s.db)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	for _, st := range stats {
		if st.Type == models.VoteTypeDifficulty && (st.Count != 1 || st.Avg != 8) {
			t.Errorf("difficulty stats = %+v, want count 1 avg 8 (latest score 4)", st)
		}
	}

	rec := httptest.NewRecorder()
	s.handleDiscipline(rec, disciplineRequest(true))
	body := rec.Body.String()
	for _, want := range []string{
		`data-type="5"
                data-score="4"`,
		`data-type="1"
                data-score="3"`,
		`data-type="2"
                data-score=""`,
		"Você já avaliou",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("logged-in discipline page lacks %q", want)
		}
	}

	// The list page average is the user's latest general vote (3 * 2).
	if !strings.Contains(body, "6.00") {
		t.Errorf("discipline page should show the average 6.00 from the latest vote")
	}

	// Logged out there is no prefill and no notice.
	rec = httptest.NewRecorder()
	s.handleDiscipline(rec, disciplineRequest(false))
	if strings.Contains(rec.Body.String(), "Você já avaliou") {
		t.Errorf("logged-out page should not mention previous votes")
	}
}

// TestVoteActivityIsNeverNull checks the heatmap endpoint returns a JSON
// array (empty when there are no votes) with day strings the page can use.
func TestVoteActivityIsNeverNull(t *testing.T) {
	s := newAuthServer(t)
	cp := seedClassProfessor(t, s.db)

	rec := httptest.NewRecorder()
	s.handleVoteActivity(rec, httptest.NewRequest(http.MethodGet, "/api/vote-activity?id=1", nil))
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("empty activity = %q, want []", rec.Body.String())
	}

	s.db.Create(&models.Vote{ClassProfessorID: cp.ID, UserID: "1", Type: 1, Score: 3, Time: time.Now().Unix()})
	rec = httptest.NewRecorder()
	s.handleVoteActivity(rec, httptest.NewRequest(http.MethodGet, "/api/vote-activity?id=1", nil))
	var activity []struct {
		Date  int64 `json:"date"`
		Value int   `json:"value"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &activity); err != nil {
		t.Fatalf("decode activity: %v (%s)", err, rec.Body.String())
	}
	if len(activity) != 1 || activity[0].Value != 1 || activity[0].Date == 0 {
		t.Errorf("activity = %+v, want one day with one vote", activity)
	}
}
