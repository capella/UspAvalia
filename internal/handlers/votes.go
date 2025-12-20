package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
)

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	classProfessorID, err := strconv.ParseUint(r.FormValue("class_professor_id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	score, err := strconv.Atoi(r.FormValue("score"))
	if err != nil || score < 1 || score > 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	voteType, err := strconv.Atoi(r.FormValue("type"))
	if err != nil {
		voteType = 1
	}

	vote := models.Vote{
		ClassProfessorID: uint(classProfessorID),
		UserID:           userID,
		Score:            score,
		Type:             voteType,
		Time:             time.Now().Unix(),
	}

	s.db.Create(&vote)
	middleware.RecordVote()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

func (s *Server) handleBatchVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request struct {
		ClassProfessorID uint `json:"class_professor_id"`
		Votes            []struct {
			Type  int `json:"type"`
			Score int `json:"score"`
		} `json:"votes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.ClassProfessorID == 0 || len(request.Votes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate votes and create batch
	var votes []models.Vote
	currentTime := time.Now().Unix()

	for _, voteData := range request.Votes {
		if voteData.Score < 0 || voteData.Score > 5 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if voteData.Type < 1 || voteData.Type > 5 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		vote := models.Vote{
			ClassProfessorID: request.ClassProfessorID,
			UserID:           userID,
			Score:            voteData.Score,
			Type:             voteData.Type,
			Time:             currentTime,
		}
		votes = append(votes, vote)
	}

	// Create all votes in a single transaction
	if err := s.db.Create(&votes).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Record metrics for each vote
	for range votes {
		middleware.RecordVote()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}