package database

import (
	"testing"
	"uspavalia/internal/models"
)

// TestLatestVotesKeepsNewestPerUserAndType checks that re-voting keeps the
// history in votes but only the newest vote per user and criterion is
// visible through LatestVotes, and therefore in ListaMedias.
func TestLatestVotesKeepsNewestPerUserAndType(t *testing.T) {
	db := newTestDB(t)
	cpID := seedClassProfessor(t, db, "MAC0110")

	votes := []models.Vote{
		{ClassProfessorID: cpID, UserID: "u1", Type: 1, Score: 1, Time: 100},
		{ClassProfessorID: cpID, UserID: "u1", Type: 1, Score: 5, Time: 200}, // newer, wins
		{ClassProfessorID: cpID, UserID: "u1", Type: 5, Score: 2, Time: 300}, // other criterion
		{ClassProfessorID: cpID, UserID: "u2", Type: 1, Score: 3, Time: 100}, // other user
	}
	if err := db.Create(&votes).Error; err != nil {
		t.Fatalf("create votes: %v", err)
	}

	var history int64
	db.Model(&models.Vote{}).Count(&history)
	if history != 4 {
		t.Fatalf("votes table should keep all %d rows, has %d", 4, history)
	}

	var latest []models.Vote
	if err := db.Table(models.LatestVotesTable).Order("user_id, type").Find(&latest).Error; err != nil {
		t.Fatalf("query latest votes: %v", err)
	}
	if len(latest) != 3 {
		t.Fatalf("LatestVotes has %d rows, want 3", len(latest))
	}
	if latest[0].UserID != "u1" || latest[0].Type != 1 || latest[0].Score != 5 {
		t.Errorf("u1 type 1 latest = %+v, want score 5", latest[0])
	}

	var media struct {
		Avg   float64 `gorm:"column:AVG(nota)"`
		Count int64   `gorm:"column:COUNT(*)"`
	}
	err := db.Raw(`SELECT "AVG(nota)", "COUNT(*)" FROM ListaMedias WHERE class_professor_id = ?`, cpID).Scan(&media).Error
	if err != nil {
		t.Fatalf("query ListaMedias: %v", err)
	}
	// u1 latest (5) and u2 (3); difficulty excluded.
	if media.Count != 2 || media.Avg != 4 {
		t.Errorf("ListaMedias count=%d avg=%v, want count 2 avg 4", media.Count, media.Avg)
	}
}
