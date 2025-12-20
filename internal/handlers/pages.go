package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
	"uspavalia/pkg/auth"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RatingStat struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Average float64 `json:"average"`
	StdDev  float64 `json:"stddev"`
}

type CommentWithVotes struct {
	models.Comment
	FormattedTime string `json:"formatted_time"`
	PositiveVotes int    `json:"positive_votes"`
	NegativeVotes int    `json:"negative_votes"`
}

func (s *Server) getCurrentUser(r *http.Request) *models.User {
	if userID, ok := middleware.GetUserID(r); ok {
		var user models.User
		if err := s.db.First(&user, userID).Error; err == nil {
			return &user
		}
	}
	return nil
}

func (s *Server) calculateStats() *models.Stats {
	var avgRating float64
	var totalEvaluations int64
	var totalUsers int64

	// Calculate average rating (excluding type 5 - difficulty)
	var result struct {
		Avg float64
	}
	if err := s.db.Raw("SELECT AVG(score) as avg FROM votes WHERE type <> 5").Scan(&result).Error; err != nil {
		logrus.Errorf("Failed to calculate average rating: %v", err)
	}
	avgRating = result.Avg * 2 // Multiply by 2 like original PHP code

	// Count total evaluations
	s.db.Model(&models.Vote{}).Count(&totalEvaluations)

	// Count unique users who voted
	var usersResult struct {
		Count int64
	}
	if err := s.db.Raw("SELECT COUNT(DISTINCT user_id) as count FROM votes").Scan(&usersResult).Error; err != nil {
		logrus.Errorf("Failed to count users: %v", err)
	}
	totalUsers = usersResult.Count

	return &models.Stats{
		AverageRating:    formatNumber(avgRating, 2),
		TotalEvaluations: formatNumber(float64(totalEvaluations), 0),
		TotalUsers:       formatNumber(float64(totalUsers), 0),
	}
}

// formatNumber formats numbers with Portuguese locale (comma as decimal separator)
func formatNumber(num float64, decimals int) string {
	var formatted string
	if decimals > 0 {
		formatted = fmt.Sprintf("%."+strconv.Itoa(decimals)+"f", num)
	} else {
		formatted = fmt.Sprintf("%.0f", num)
	}
	// Replace dot with comma for Portuguese formatting
	formatted = strings.Replace(formatted, ".", ",", 1)
	return formatted
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	var bestRated []models.BestRated

	// Try to query the view, fallback to empty if view doesn't exist
	result := s.db.Limit(10).Find(&bestRated)
	if result.Error != nil {
		logrus.Printf("Warning: Could not load best rated data: %v", result.Error)
		bestRated = []models.BestRated{} // Empty slice as fallback
	}

	// Calculate statistics
	stats := s.calculateStats()

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"BestRated": bestRated,
			"Stats":     stats,
			"BaseURL":   fmt.Sprintf("http://%s", r.Host),
		},
	}

	s.renderTemplate(w, r, "index", data)
}

func (s *Server) handleDiscipline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var discipline models.Discipline
	if err := s.db.Preload("Unit").First(&discipline, id).Error; err != nil {
		s.renderErrorPage(w, r, http.StatusNotFound, "Disciplina não encontrada")
		return
	}

	// Use a custom query with LEFT JOIN to calculate media in one query
	var results []struct {
		models.ClassProfessor
		Media *float64 `gorm:"column:media"`
	}

	err = s.db.Model(&models.ClassProfessor{}).
		Select("class_professors.*, COALESCE(AVG(votes.score) * 2, 0) as media").
		Joins("LEFT JOIN votes ON votes.class_professor_id = class_professors.id AND votes.type <> 5").
		Preload("Professor").
		Where("class_professors.class_id = ?", id).
		Group("class_professors.id").
		Find(&results).Error

	if err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao carregar dados")
		return
	}

	modals := []map[string]interface{}{}
	for _, result := range results {
		mediaValue := 0.0
		if result.Media != nil {
			mediaValue = *result.Media
		}
		modals = append(modals, map[string]interface{}{
			"ClassProfessor": result.ClassProfessor,
			"Media":          mediaValue,
			"CSRFToken":      csrf.Token(r),
		})
	}

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"Discipline": discipline,
			"List":       modals,
		},
	}

	s.renderTemplate(
		w,
		r,
		"disciplina",
		data,
		"templates/vote-modal.html",
		"templates/thank-you-modal.html",
	)
}

func (s *Server) handleProfessor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var professor models.Professor
	if err := s.db.Preload("Unit").First(&professor, id).Error; err != nil {
		s.renderErrorPage(w, r, http.StatusNotFound, "Professor não encontrado")
		return
	}

	// Use a custom query with LEFT JOIN to calculate media in one query
	var results []struct {
		models.ClassProfessor
		Media *float64 `gorm:"column:media"`
	}

	err = s.db.Model(&models.ClassProfessor{}).
		Select("class_professors.*, COALESCE(AVG(votes.score) * 2, 0) as media").
		Joins("LEFT JOIN votes ON votes.class_professor_id = class_professors.id AND votes.type <> 5").
		Preload("Discipline").
		Where("class_professors.professor_id = ?", id).
		Group("class_professors.id").
		Find(&results).Error

	if err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao carregar dados")
		return
	}

	modals := []map[string]interface{}{}
	for _, result := range results {
		mediaValue := 0.0
		if result.Media != nil {
			mediaValue = *result.Media
		}
		modals = append(modals, map[string]interface{}{
			"ClassProfessor": result.ClassProfessor,
			"Media":          mediaValue,
			"CSRFToken":      csrf.Token(r),
		})
	}

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"Professor": professor,
			"List":      modals,
		},
	}

	s.renderTemplate(
		w,
		r,
		"professor",
		data,
		"templates/vote-modal.html",
		"templates/thank-you-modal.html",
	)
}

func (s *Server) handleVer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get ClassProfessor with related Professor and Discipline
	var classProfessor models.ClassProfessor
	if err := s.db.Preload("Professor.Unit").Preload("Discipline").First(&classProfessor, id).Error; err != nil {
		s.renderErrorPage(w, r, http.StatusNotFound, "Avaliação não encontrada")
		return
	}

	// Get rating statistics using CalculateStatsByType method
	stats, err := classProfessor.CalculateStatsByType(s.db)
	if err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao calcular estatísticas")
		return
	}

	ratingNames := map[models.VoteType]string{
		models.VoteTypeGeneral:         "Avaliação Geral",
		models.VoteTypeTeaching:        "Didática",
		models.VoteTypeCommitment:      "Empenho/Dedicação",
		models.VoteTypeStudentRelation: "Relação com os alunos",
		models.VoteTypeDifficulty:      "Dificuldade",
	}

	var ratingStats []RatingStat
	var totalVotes int64
	for _, stat := range stats {
		ratingStats = append(ratingStats, RatingStat{
			Name:    ratingNames[stat.Type],
			Count:   int(stat.Count),
			Average: stat.Avg,
			StdDev:  stat.Std,
		})
		totalVotes += stat.Count
	}

	// Get comments with vote counts
	var comments []CommentWithVotes
	query := `
		SELECT
			c.*,
			COALESCE(pos.positive_votes, 0) as positive_votes,
			COALESCE(neg.negative_votes, 0) as negative_votes
		FROM comments c
		LEFT JOIN (
			SELECT comment_id, COUNT(*) as positive_votes
			FROM comment_votes
			WHERE vote = 1
			GROUP BY comment_id
		) pos ON c.id = pos.comment_id
		LEFT JOIN (
			SELECT comment_id, COUNT(*) as negative_votes
			FROM comment_votes
			WHERE vote = -1
			GROUP BY comment_id
		) neg ON c.id = neg.comment_id
		WHERE class_professor_id = ?
		AND (COALESCE(pos.positive_votes, 0) - COALESCE(neg.negative_votes, 0)) >= -3
		ORDER BY time DESC`

	rows, err := s.db.Raw(query, id).Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var comment CommentWithVotes
			var rawTime int64

			err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content,
				&comment.ClassProfessorID, &rawTime, &comment.CreatedAt, &comment.UpdatedAt,
				&comment.PositiveVotes, &comment.NegativeVotes)
			if err == nil {
				comment.Time = rawTime
				comment.FormattedTime = time.Unix(rawTime, 0).Format("02/01/2006 - 15:04:05")
				comments = append(comments, comment)
			}
		}
	}

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"ClassProfessor": classProfessor,
			"Professor":      classProfessor.Professor,
			"Discipline":     classProfessor.Discipline,
			"RatingStats":    ratingStats,
			"Comments":       comments,
			"TotalVotes":     totalVotes,
			"Modal": map[string]interface{}{
				"CSRFToken":      csrf.TemplateField(r),
				"ClassProfessor": classProfessor,
			},
		},
	}

	s.renderTemplate(
		w,
		r,
		"ver",
		data,
		"templates/vote-modal.html",
		"templates/thank-you-modal.html",
	)
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
	}
	s.renderTemplate(w, r, "sobre", data)
}

func (s *Server) handleContact(w http.ResponseWriter, r *http.Request) {
	currentUser := s.getCurrentUser(r)

	if r.Method == "GET" {
		data := PageData{
			CSRFToken: csrf.Token(r),
			User:      currentUser,
			Data: map[string]interface{}{
				"HCaptchaSiteKey": s.config.Security.HCaptchaSiteKey,
			},
		}
		s.renderTemplate(w, r, "contact", data)
		return
	}

	// POST request - handle form submission
	if err := r.ParseForm(); err != nil {
		s.renderContactError(w, r, "Dados do formulário inválidos")
		return
	}

	firstName := strings.TrimSpace(r.FormValue("first_name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	telephone := strings.TrimSpace(r.FormValue("telephone"))
	comments := strings.TrimSpace(r.FormValue("comments"))

	// Validate required fields
	if firstName == "" || lastName == "" || email == "" || comments == "" {
		s.renderContactError(w, r, "Todos os campos obrigatórios devem ser preenchidos")
		return
	}

	// Verify hCaptcha for non-logged-in users
	if currentUser == nil {
		hcaptchaResponse := r.FormValue("h-captcha-response")
		if hcaptchaResponse == "" {
			s.renderContactError(w, r, "Por favor, complete o desafio de segurança")
			return
		}

		valid, err := auth.VerifyHCaptcha(
			s.config.Security.HCaptchaSecretKey,
			hcaptchaResponse,
			r.RemoteAddr,
		)
		if err != nil || !valid {
			logrus.Printf("hCaptcha verification failed: %v", err)
			s.renderContactError(w, r, "Verificação de segurança falhou. Por favor, tente novamente")
			return
		}
	}

	// Validate email
	if !validateEmail(email) {
		s.renderContactError(w, r, "Email inválido")
		return
	}

	// Validate name
	nameRegex := regexp.MustCompile(`^[A-Za-zÀ-ÿ\s.'-]+$`)
	if !nameRegex.MatchString(firstName) || !nameRegex.MatchString(lastName) {
		s.renderContactError(w, r, "Nome inválido")
		return
	}

	// Validate comments length
	if len(comments) < 2 {
		s.renderContactError(w, r, "Mensagem muito curta")
		return
	}

	// Send email
	if err := s.emailService.SendContactEmail(firstName, lastName, email, telephone, comments); err != nil {
		logrus.Printf("Contact email error: %v", err)
		s.renderContactError(w, r, "Erro ao enviar mensagem. Por favor, tente novamente mais tarde")
		return
	}

	// Show success
	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"Success": "Mensagem enviada com sucesso! Obrigado pelo contato.",
		},
	}
	s.renderTemplate(w, r, "contact", data)
}

func (s *Server) renderContactError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	currentUser := s.getCurrentUser(r)

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      currentUser,
		Data: map[string]interface{}{
			"Error":           errorMsg,
			"HCaptchaSiteKey": s.config.Security.HCaptchaSiteKey,
		},
	}
	s.renderTemplate(w, r, "contact", data)
}

func (s *Server) handleTopRated(w http.ResponseWriter, r *http.Request) {
	// Get best rated disciplines (from the Melhores view)
	var bestRatedDisciplines []models.BestRated
	s.db.Limit(10).Find(&bestRatedDisciplines)

	// Get best rated professors separately
	type BestRatedProfessor struct {
		ProfessorID   uint    `json:"professor_id"`
		ProfessorName string  `json:"professor_name"`
		UnitName      string  `json:"unit_name"`
		Average       float64 `json:"average"`
		VoteCount     int     `json:"vote_count"`
	}

	var bestRatedProfessors []BestRatedProfessor
	s.db.Raw(`
		SELECT
			p.id as professor_id,
			p.name as professor_name,
			u.name as unit_name,
			(AVG(v.score) * 2) as average,
			COUNT(*) as vote_count
		FROM votes v
		INNER JOIN class_professors ap ON v.class_professor_id = ap.id
		INNER JOIN professors p ON ap.professor_id = p.id
		INNER JOIN units u ON p.unit_id = u.id
		WHERE v.type <> 5
		GROUP BY p.id, p.name, u.name
		HAVING COUNT(*) >= 15
		ORDER BY AVG(v.score) DESC, COUNT(*) DESC
		LIMIT 10
	`).Scan(&bestRatedProfessors)

	data := PageData{
		CSRFToken: csrf.Token(r),
		User:      s.getCurrentUser(r),
		Data: map[string]interface{}{
			"BestRatedDisciplines": bestRatedDisciplines,
			"BestRatedProfessors":  bestRatedProfessors,
		},
	}

	s.renderTemplate(w, r, "10melhores", data)
}
