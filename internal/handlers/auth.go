package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
	"uspavalia/pkg/auth"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type PageData struct {
	User *models.User
	Data interface{}
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateEmail validates email format
func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// loginNextSessionKey stores where to send the user once they finish logging in.
const loginNextSessionKey = "login_next"

// safeNextPath returns raw if it is a path on this site (e.g.
// "/disciplina/1#modal2"), or "" for anything else, so a crafted link can
// never turn the login flow into an open redirect.
func safeNextPath(raw string) string {
	if raw == "" || raw[0] != '/' || strings.ContainsAny(raw, "\\\r\n") {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil ||
		!strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return ""
	}
	return u.String()
}

// rememberLoginNext records the page (and modal) to return to after login.
// Without a valid "next" any stale value is cleared.
func (s *Server) rememberLoginNext(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	if next := safeNextPath(r.URL.Query().Get("next")); next != "" {
		session.Values[loginNextSessionKey] = next
	} else {
		delete(session.Values, loginNextSessionKey)
	}
	session.Save(r, w)
}

// popLoginNext consumes the remembered return path, defaulting to home.
func popLoginNext(session *sessions.Session) string {
	next, _ := session.Values[loginNextSessionKey].(string)
	delete(session.Values, loginNextSessionKey)
	if next = safeNextPath(next); next != "" {
		return next
	}
	return "/"
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
	// A "next" given directly to this route wins over the one remembered by
	// the login page; otherwise keep whatever the login page stored.
	if next := safeNextPath(r.URL.Query().Get("next")); next != "" {
		session.Values[loginNextSessionKey] = next
	}
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
			EmailHash: emailHash,
		}
		if err := s.db.Create(&user).Error; err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		middleware.RecordRegistration()
	}

	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	delete(session.Values, "oauth_state")
	next := popLoginNext(session)
	session.Save(r, w)

	http.Redirect(w, r, next, http.StatusFound)
}

// handleRequestLogin shows the magic link request form and processes login requests
func (s *Server) handleRequestLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.rememberLoginNext(w, r)
		data := PageData{
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

	if hcaptchaResponse != "" {
		valid, err := auth.VerifyHCaptcha(
			s.config.Security.HCaptchaSecretKey,
			hcaptchaResponse,
			r,
		)
		if err != nil || !valid {
			if s.config.DevMode {
				logrus.Printf(
					"[DEV MODE] hCaptcha verification failed: %v - bypassing validation",
					err,
				)
			} else {
				logrus.Printf("hCaptcha verification failed: %v", err)
				s.renderLoginRequestError(w, r, "Verificação de segurança falhou")
				return
			}
		}
	}

	// Validate email
	if !validateEmail(email) {
		s.renderLoginRequestError(w, r, "Email inválido")
		return
	}

	// Generate email hash
	emailHash := auth.HashEmail(email, s.config.Security.SecretKey)

	// Generate one-time use token
	token := auth.GenerateSecureToken()
	expiresAt := time.Now().Add(auth.MagicLinkExpiry)

	// Store token in database
	loginToken := models.LoginToken{
		Token:     token,
		EmailHash: emailHash,
		ExpiresAt: expiresAt,
	}
	if err := s.db.Create(&loginToken).Error; err != nil {
		logrus.Printf("Error creating login token: %v", err)
		s.renderLoginRequestError(w, r, "Erro ao gerar link de login. Tente novamente.")
		return
	}

	// Create login URL
	loginURL := fmt.Sprintf("%s/auth/magic-link?token=%s", s.config.Server.URL, token)

	// Log magic link in dev mode
	if s.config.DevMode {
		logrus.Printf("[DEV MODE] Magic link for %s: %s", email, loginURL)
	}

	// Send email
	if err := s.emailService.SendMagicLink(email, loginURL); err != nil {
		logrus.Printf("Error sending magic link email: %v", err)
		s.renderLoginRequestError(w, r, "Erro ao enviar email. Tente novamente.")
		return
	}

	s.renderLoginSent(w, r)
}

// handleMagicLink processes magic link authentication
func (s *Server) handleMagicLink(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")

	if tokenStr == "" {
		s.renderErrorPage(w, r, http.StatusBadRequest, "Link inválido")
		return
	}

	// Look up token in database
	var loginToken models.LoginToken
	if err := s.db.Where("token = ?", tokenStr).First(&loginToken).Error; err != nil {
		s.renderErrorPage(w, r, http.StatusUnauthorized, "Link de login inválido ou já utilizado")
		return
	}

	// Check if token is expired
	if time.Now().After(loginToken.ExpiresAt) {
		// Delete expired token
		s.db.Delete(&loginToken)
		s.renderErrorPage(w, r, http.StatusUnauthorized, "Link de login expirado")
		return
	}

	// Delete token immediately (one-time use)
	if err := s.db.Delete(&loginToken).Error; err != nil {
		logrus.Printf("Error deleting login token: %v", err)
		// Continue anyway - the login should still work
	}

	// Find or create user
	var user models.User
	result := s.db.Where("email_hash = ?", loginToken.EmailHash).First(&user)
	if result.Error != nil {
		// Create new user
		user = models.User{
			EmailHash: loginToken.EmailHash,
		}
		if err := s.db.Create(&user).Error; err != nil {
			logrus.Printf("Error creating user: %v", err)
			s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao criar usuário")
			return
		}
		middleware.RecordRegistration()
	}

	// Create session
	session, _ := s.store.Get(r, s.config.Security.SessionName)
	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	next := popLoginNext(session)
	if err := session.Save(r, w); err != nil {
		logrus.Printf("Error saving session: %v", err)
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Erro ao criar sessão")
		return
	}

	// Back to the page (and modal) the user came from, or home
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// renderLoginRequestError renders the login request page with an error message
func (s *Server) renderLoginRequestError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := PageData{
		Data: map[string]interface{}{
			"Error":           errorMsg,
			"HCaptchaSiteKey": s.config.Security.HCaptchaSiteKey,
		},
	}
	s.renderTemplate(w, r, "request-login", data)
}

// renderLoginSent renders the "login email sent" confirmation page
func (s *Server) renderLoginSent(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	s.renderTemplate(w, r, "login-sent", data)
}
