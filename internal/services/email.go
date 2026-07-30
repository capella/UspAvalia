package services

import (
	"bytes"
	"context"
	"fmt"
	htmlTemplate "html/template"
	"log"
	textTemplate "text/template"
	"time"
	"uspavalia/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type EmailService struct {
	config        *config.Config
	client        *sesv2.Client
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
		client:        newSESClient(cfg),
		htmlTemplates: htmlTemplates,
		textTemplates: textTemplates,
	}
}

func newSESClient(cfg *config.Config) *sesv2.Client {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Email.AWSRegion),
	}
	if cfg.Email.AWSAccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.Email.AWSAccessKeyID,
				cfg.Email.AWSSecretAccessKey,
				"",
			),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		log.Printf("Warning: Failed to load AWS config for SES: %v", err)
		return nil
	}

	return sesv2.NewFromConfig(awsCfg)
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

// SendEmail sends an email using AWS SES
func (es *EmailService) SendEmail(toEmail, toName string, template EmailTemplate) error {
	if es.client == nil {
		return fmt.Errorf("email service not configured")
	}

	from := es.config.Email.FromEmail
	if es.config.Email.FromName != "" {
		from = fmt.Sprintf("%s <%s>", es.config.Email.FromName, es.config.Email.FromEmail)
	}

	utf8 := func(s string) *types.Content {
		return &types.Content{Data: aws.String(s), Charset: aws.String("UTF-8")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := es.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: utf8(template.Subject),
				Body: &types.Body{
					Text: utf8(template.PlainText),
					Html: utf8(template.HTMLContent),
				},
			},
		},
	})
	if err != nil {
		log.Printf("SES error: %v", err)
		return fmt.Errorf("email service error: %w", err)
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
