package services

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"log"
	textTemplate "text/template"
	"uspavalia/internal/config"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
	config        *config.Config
	apiKey        string
	htmlTemplates *htmlTemplate.Template
	textTemplates *textTemplate.Template
}

type EmailTemplate struct {
	Subject     string
	HTMLContent string
	PlainText   string
}

func NewEmailService(cfg *config.Config) *EmailService {
	// Load HTML templates
	htmlTemplates, err := htmlTemplate.ParseGlob("templates/emails/*.html")
	if err != nil {
		log.Printf("Warning: Failed to load HTML email templates: %v", err)
	}

	// Load text templates
	textTemplates, err := textTemplate.ParseGlob("templates/emails/*.txt")
	if err != nil {
		log.Printf("Warning: Failed to load text email templates: %v", err)
	}

	return &EmailService{
		config:        cfg,
		apiKey:        cfg.Email.SendGridAPIKey,
		htmlTemplates: htmlTemplates,
		textTemplates: textTemplates,
	}
}

// renderTemplate renders both HTML and text versions of an email template
func (es *EmailService) renderTemplate(templateName string, data interface{}) (htmlContent, plainText string, err error) {
	// Render HTML template
	htmlBuf := new(bytes.Buffer)
	htmlTemplateName := templateName + ".html"
	if err := es.htmlTemplates.ExecuteTemplate(htmlBuf, htmlTemplateName, data); err != nil {
		return "", "", fmt.Errorf("failed to render HTML template %s: %w", htmlTemplateName, err)
	}
	htmlContent = htmlBuf.String()

	// Render text template
	textBuf := new(bytes.Buffer)
	textTemplateName := templateName + ".txt"
	if err := es.textTemplates.ExecuteTemplate(textBuf, textTemplateName, data); err != nil {
		return "", "", fmt.Errorf("failed to render text template %s: %w", textTemplateName, err)
	}
	plainText = textBuf.String()

	return htmlContent, plainText, nil
}

// SendEmail sends an email using SendGrid
func (es *EmailService) SendEmail(toEmail, toName string, template EmailTemplate) error {
	from := mail.NewEmail(es.config.Email.FromName, es.config.Email.FromEmail)
	to := mail.NewEmail(toName, toEmail)

	message := mail.NewSingleEmail(
		from,
		template.Subject,
		to,
		template.PlainText,
		template.HTMLContent,
	)

	client := sendgrid.NewSendClient(es.apiKey)
	response, err := client.Send(message)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	if response.StatusCode >= 400 {
		log.Printf("SendGrid error: Status %d, Body: %s", response.StatusCode, response.Body)
		return fmt.Errorf("email service error: %d", response.StatusCode)
	}

	if es.config.DevMode {
		log.Printf("Email sent successfully (dev mode)\nSubject: %s\nPlain Text:\n%s\nHTML:\n%s",
			template.Subject,
			template.PlainText,
			template.HTMLContent,
		)
	} else {
		log.Printf("Email sent successfully")
	}
	return nil
}

// SendContactEmail sends a contact form submission to the admin
func (es *EmailService) SendContactEmail(
	firstName, lastName, email, comments string,
) error {
	adminEmail := "contato@uspavalia.com"
	if es.config.Email.FromEmail != "" {
		adminEmail = es.config.Email.FromEmail
	}

	data := map[string]string{
		"FirstName": firstName,
		"LastName":  lastName,
		"Email":     email,
		"Comments":  comments,
	}

	htmlContent, plainText, err := es.renderTemplate("contact", data)
	if err != nil {
		return err
	}

	template := EmailTemplate{
		Subject:     "USP Avalia - Contato",
		HTMLContent: htmlContent,
		PlainText:   plainText,
	}

	return es.SendEmail(adminEmail, "USP Avalia Admin", template)
}

// SendMagicLink sends a magic link authentication email
func (es *EmailService) SendMagicLink(toEmail, loginURL string) error {
	data := map[string]string{
		"LoginURL": loginURL,
	}

	htmlContent, plainText, err := es.renderTemplate("magic-link", data)
	if err != nil {
		return err
	}

	template := EmailTemplate{
		Subject:     "Seu link de login - UspAvalia",
		HTMLContent: htmlContent,
		PlainText:   plainText,
	}

	return es.SendEmail(toEmail, "", template)
}
