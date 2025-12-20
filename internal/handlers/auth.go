package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
	"uspavalia/pkg/auth"

	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type PageData struct {
	CSRFToken         string
	CSRFTokenTemplate template.HTML
	User              *models.User
	Data              interface{}
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateEmail validates email format
func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// getUserByEmailHash finds user by email hash
func (s *Server) getUserByEmailHash(email string) (*models.User, error) {
	emailHash := auth.HashEmail(email, s.config.Security.SecretKey)

	var user models.User
	if err := s.db.Where("email_hash = ?", emailHash).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Server) handleEmailVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		s.renderErrorPage(w, r, http.StatusBadRequest, "Token de verificação inválido")
		return
	}

	var user models.User
	if err := s.db.Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		s.renderErrorPage(
			w,
			r,
			http.StatusNotFound,
			"Token de verificação não encontrado ou já usado",
		)
		return
	}

	if auth.IsTokenExpired(user.EmailVerificationExpiry) {
		s.renderErrorPage(w, r, http.StatusBadRequest, "Token de verificação expirado")
		return
	}

	// Verify email
	user.EmailVerified = true
	user.EmailVerificationToken = ""
	user.EmailVerificationExpiry = nil

	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("Email verification save error: %v", err)
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao verificar email")
		return
	}

	// Auto-login the user
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	session.Save(r, w)

	data := PageData{
		CSRFToken: csrf.Token(r),
		Data: map[string]interface{}{
			"Success": "Email verificado com sucesso! Você está agora logado.",
		},
	}
	s.renderTemplate(w, r, "email-verified", data)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	delete(session.Values, "user_id")
	session.Save(r, w)

	// Redirect to the previous page (referer) or home page as fallback
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	http.Redirect(w, r, referer, http.StatusFound)
}

func (s *Server) getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.OAuth.Google.ClientID,
		ClientSecret: s.config.OAuth.Google.ClientSecret,
		RedirectURL:  s.config.OAuth.Google.RedirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	config := s.getGoogleOAuthConfig()
	state := auth.GenerateRandomString(32)

	session, _ := s.store.Get(r, s.config.Security.SessionName)
	session.Values["oauth_state"] = state
	session.Save(r, w)

	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || r.URL.Query().Get("state") != expectedState {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	config := s.getGoogleOAuthConfig()
	token, err := config.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	var user models.User
	// Try to find by email hash
	emailHash := auth.HashEmail(googleUser.Email, s.config.Security.SecretKey)
	result := s.db.Where("email_hash = ?", emailHash).First(&user)
	if result.Error != nil {
		// Create new user
		user = models.User{
			EmailHash:     emailHash,
			EmailVerified: true, // Google emails are pre-verified
		}
		if err := s.db.Create(&user).Error; err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		middleware.RecordRegistration()
	} else {
		// Link Google account
		user.EmailVerified = true // Google emails are pre-verified
		s.db.Save(&user)
	}

	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	delete(session.Values, "oauth_state")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// handleRequestLogin shows the magic link request form and processes login requests
func (s *Server) handleRequestLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			CSRFToken: csrf.Token(r),
			Data: map[string]interface{}{
				"HCaptchaSiteKey": s.config.Security.HCaptchaSiteKey,
			},
		}
		s.renderTemplate(w, r, "request-login", data)
		return
	}

	// POST - process magic link request
	if err := r.ParseForm(); err != nil {
		s.renderLoginRequestError(w, r, "Dados do formulário inválidos")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" {
		s.renderLoginRequestError(w, r, "Por favor, informe seu email")
		return
	}

	// Verify hCaptcha
	hcaptchaResponse := r.FormValue("h-captcha-response")
	if hcaptchaResponse == "" {
		s.renderLoginRequestError(w, r, "Por favor, complete o desafio de segurança")
		return
	}

	valid, err := auth.VerifyHCaptcha(
		s.config.Security.HCaptchaSecretKey,
		hcaptchaResponse,
		r.RemoteAddr,
	)
	if err != nil || !valid {
		log.Printf("hCaptcha verification failed: %v", err)
		s.renderLoginRequestError(w, r, "Verificação de segurança falhou")
		return
	}

	// Validate email
	if !validateEmail(email) {
		s.renderLoginRequestError(w, r, "Email inválido")
		return
	}

	// Find user by email hash
	emailHash := auth.HashEmail(email, s.config.Security.SecretKey)
	var user models.User
	if err := s.db.Where("email_hash = ?", emailHash).First(&user).Error; err != nil {
		// Don't reveal if user exists - always show success
		s.renderLoginSent(w, r, email)
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		s.renderLoginRequestError(w, r, "Email não verificado. Verifique seu email primeiro.")
		return
	}

	// Generate magic link token
	token, _ := auth.GenerateMagicLinkToken(emailHash, []byte(s.config.Security.MagicLinkHMACKey))

	// Create login URL
	loginURL := fmt.Sprintf("%s/auth/magic-link?token=%s", s.config.Server.URL, token)

	// Send email
	if err := s.emailService.SendMagicLink(email, loginURL); err != nil {
		log.Printf("Error sending magic link email: %v", err)
		s.renderLoginRequestError(w, r, "Erro ao enviar email. Tente novamente.")
		return
	}

	s.renderLoginSent(w, r, email)
}

// handleMagicLink processes magic link authentication
func (s *Server) handleMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		s.renderErrorPage(w, r, http.StatusBadRequest, "Link inválido")
		return
	}

	// Verify token
	emailHash, valid := auth.VerifyMagicLinkToken(token, []byte(s.config.Security.MagicLinkHMACKey))
	if !valid {
		s.renderErrorPage(w, r, http.StatusUnauthorized, "Link de login expirado ou inválido")
		return
	}

	// Find user
	var user models.User
	if err := s.db.Where("email_hash = ? AND email_verified = ?", emailHash, true).First(&user).Error; err != nil {
		s.renderErrorPage(w, r, http.StatusNotFound, "Usuário não encontrado ou email não verificado")
		return
	}

	// Create session
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao criar sessão")
		return
	}

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderLoginRequestError renders the login request page with an error message
func (s *Server) renderLoginRequestError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := PageData{
		CSRFToken: csrf.Token(r),
		Data: map[string]interface{}{
			"Error":           errorMsg,
			"HCaptchaSiteKey": s.config.Security.HCaptchaSiteKey,
		},
	}
	s.renderTemplate(w, r, "request-login", data)
}

// renderLoginSent renders the "login email sent" confirmation page
func (s *Server) renderLoginSent(w http.ResponseWriter, r *http.Request, email string) {
	data := PageData{
		CSRFToken: csrf.Token(r),
		Data: map[string]interface{}{
			"Email": email,
		},
	}
	s.renderTemplate(w, r, "login-sent", data)
}
