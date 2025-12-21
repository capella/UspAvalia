package services

import (
	"fmt"
	"log"
	"uspavalia/internal/config"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
	config *config.Config
	apiKey string
}

type EmailTemplate struct {
	Subject     string
	HTMLContent string
	PlainText   string
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
		apiKey: cfg.Email.SendGridAPIKey,
	}
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

// SendVerificationEmail sends an email verification email
func (es *EmailService) SendVerificationEmail(toEmail, toName, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", es.config.Server.URL, token)

	template := EmailTemplate{
		Subject: "Verificação de Email - USP Avalia",
		PlainText: fmt.Sprintf(`
Olá %s,

Obrigado por se registrar no USP Avalia!

Para verificar seu email, clique no link abaixo ou copie e cole no seu navegador:
%s

Este link expira em 24 horas.

Se você não criou uma conta no USP Avalia, pode ignorar este email.

Atenciosamente,
Equipe USP Avalia
		`, toName, verificationURL),
		HTMLContent: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verificação de Email - USP Avalia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; text-align: center; padding: 20px; }
        .content { padding: 30px; background-color: #f9f9f9; }
        .button { display: inline-block; background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>USP Avalia</h1>
        </div>
        <div class="content">
            <h2>Bem-vindo, %s!</h2>
            <p>Obrigado por se registrar no USP Avalia. Para completar seu cadastro, precisamos verificar seu endereço de email.</p>
            <p>Clique no botão abaixo para verificar sua conta:</p>
            <a href="%s" class="button">Verificar Email</a>
            <p>Ou copie e cole este link no seu navegador:</p>
            <p>%s</p>
            <p><strong>Este link expira em 24 horas.</strong></p>
            <p>Se você não criou uma conta no USP Avalia, pode ignorar este email.</p>
        </div>
        <div class="footer">
            <p>© 2025 USP Avalia. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
		`, toName, verificationURL, verificationURL),
	}

	return es.SendEmail(toEmail, toName, template)
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(toEmail, toName, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", es.config.Server.URL, token)

	template := EmailTemplate{
		Subject: "Redefinir Senha - USP Avalia",
		PlainText: fmt.Sprintf(`
Olá %s,

Você solicitou a redefinição de sua senha no USP Avalia.

Para redefinir sua senha, clique no link abaixo ou copie e cole no seu navegador:
%s

Este link expira em 15 minutos por segurança.

Se você não solicitou esta redefinição, pode ignorar este email.

Atenciosamente,
Equipe USP Avalia
		`, toName, resetURL),
		HTMLContent: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Redefinir Senha - USP Avalia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; text-align: center; padding: 20px; }
        .content { padding: 30px; background-color: #f9f9f9; }
        .button { display: inline-block; background-color: #dc3545; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 10px; border-radius: 4px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>USP Avalia</h1>
        </div>
        <div class="content">
            <h2>Redefinir Senha</h2>
            <p>Olá %s,</p>
            <p>Você solicitou a redefinição de sua senha no USP Avalia.</p>
            <p>Clique no botão abaixo para redefinir sua senha:</p>
            <a href="%s" class="button">Redefinir Senha</a>
            <p>Ou copie e cole este link no seu navegador:</p>
            <p>%s</p>
            <div class="warning">
                <strong>⚠️ Este link expira em 15 minutos por segurança.</strong>
            </div>
            <p>Se você não solicitou esta redefinição, pode ignorar este email com segurança.</p>
        </div>
        <div class="footer">
            <p>© 2025 USP Avalia. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
		`, toName, resetURL, resetURL),
	}

	return es.SendEmail(toEmail, toName, template)
}

// SendContactEmail sends a contact form submission to the admin
func (es *EmailService) SendContactEmail(
	firstName, lastName, email, comments string,
) error {
	adminEmail := "contato@uspavalia.com"
	if es.config.Email.FromEmail != "" {
		adminEmail = es.config.Email.FromEmail
	}

	template := EmailTemplate{
		Subject: "USP Avalia - Contato",
		PlainText: fmt.Sprintf(`
Nova mensagem de contato recebida:

Nome: %s %s
Email: %s

Mensagem:
%s

---
Enviado através do formulário de contato do USP Avalia
		`, firstName, lastName, email, comments),
		HTMLContent: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Nova Mensagem de Contato - USP Avalia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; text-align: center; padding: 20px; }
        .content { padding: 30px; background-color: #f9f9f9; }
        .field { margin-bottom: 15px; }
        .field-label { font-weight: bold; color: #666; }
        .field-value { margin-top: 5px; }
        .message-box { background-color: white; padding: 15px; border-left: 4px solid #007bff; margin-top: 10px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Nova Mensagem de Contato</h1>
        </div>
        <div class="content">
            <div class="field">
                <div class="field-label">Nome:</div>
                <div class="field-value">%s %s</div>
            </div>
            <div class="field">
                <div class="field-label">Email:</div>
                <div class="field-value"><a href="mailto:%s">%s</a></div>
            </div>
            <div class="field">
                <div class="field-label">Mensagem:</div>
                <div class="message-box">%s</div>
            </div>
        </div>
        <div class="footer">
            <p>Enviado através do formulário de contato do USP Avalia</p>
        </div>
    </div>
</body>
</html>
		`, firstName, lastName, email, email, comments),
	}

	return es.SendEmail(adminEmail, "USP Avalia Admin", template)
}

// SendMagicLink sends a magic link authentication email
func (es *EmailService) SendMagicLink(toEmail, loginURL string) error {
	template := EmailTemplate{
		Subject: "Seu link de login - UspAvalia",
		PlainText: fmt.Sprintf(`
Login UspAvalia

Clique no link abaixo para fazer login:
%s

Este link expira em 15 minutos.

Se você não solicitou este login, ignore este email.

Atenciosamente,
Equipe USP Avalia
		`, loginURL),
		HTMLContent: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Seu link de login - UspAvalia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; text-align: center; padding: 20px; }
        .content { padding: 30px; background-color: #f9f9f9; }
        .button { display: inline-block; background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 10px; border-radius: 4px; margin: 15px 0; }
        .link-box { background-color: #e9ecef; padding: 10px; border-radius: 4px; word-break: break-all; font-family: monospace; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>USP Avalia</h1>
        </div>
        <div class="content">
            <h2>Login UspAvalia</h2>
            <p>Clique no botão abaixo para fazer login:</p>
            <a href="%s" class="button">Fazer Login</a>
            <p>Ou copie e cole este link no navegador:</p>
            <div class="link-box">%s</div>
            <div class="warning">
                <strong>⚠️ Este link expira em 15 minutos.</strong>
            </div>
            <p><small>Se você não solicitou este login, ignore este email.</small></p>
        </div>
        <div class="footer">
            <p>© 2025 USP Avalia. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
		`, loginURL, loginURL),
	}

	return es.SendEmail(toEmail, "", template)
}
