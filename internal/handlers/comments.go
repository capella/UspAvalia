package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
)

func (s *Server) handleComment(w http.ResponseWriter, r *http.Request) {
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

	content := r.FormValue("comentario")
	if content == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		UserID:           userID,
		Content:          content,
		ClassProfessorID: uint(classProfessorID),
		Time:             time.Now().Unix(),
	}

	if err := s.db.Create(&comment).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	middleware.RecordComment()

	http.Redirect(w, r, fmt.Sprintf("/ver/%d", classProfessorID), http.StatusFound)
}

func (s *Server) handleCommentVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseUint(r.FormValue("comment_id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vote, err := strconv.Atoi(r.FormValue("vote"))
	if err != nil || (vote != -1 && vote != 1) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Use FirstOrCreate to find existing vote or create new one
	var commentVote models.CommentVote
	result := s.db.Where(models.CommentVote{
		CommentID: uint(commentID),
		UserID:    userID,
	}).Attrs(models.CommentVote{
		Vote: vote,
		Time: time.Now().Unix(),
	}).FirstOrCreate(&commentVote)

	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If record already existed, update the vote and time
	if result.RowsAffected == 0 {
		commentVote.Vote = vote
		commentVote.Time = time.Now().Unix()
		if err := s.db.Save(&commentVote).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Get the class professor ID from the comment to redirect back to the correct page
	var comment models.Comment
	if err := s.db.First(&comment, commentID).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ver/%d", comment.ClassProfessorID), http.StatusFound)
}