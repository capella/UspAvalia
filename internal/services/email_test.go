package services

import (
	"context"
	"os"
	"testing"
	"uspavalia/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type fakeSES struct {
	input *sesv2.SendEmailInput
}

func (f *fakeSES) SendEmail(_ context.Context, params *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.input = params
	return &sesv2.SendEmailOutput{}, nil
}

// newTestEmailService builds an EmailService with real templates (loaded
// relative to the repository root) and a fake SES sender.
func newTestEmailService(t *testing.T, cfg *config.Config) (*EmailService, *fakeSES) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	fake := &fakeSES{}
	es := NewEmailService(cfg)
	es.client = fake
	if es.htmlTemplates == nil || es.textTemplates == nil {
		t.Fatal("email templates were not loaded from templates/emails")
	}
	return es, fake
}

func TestContactRecipient(t *testing.T) {
	cases := []struct {
		name         string
		contactEmail string
		fromEmail    string
		want         string
	}{
		{"contact_email wins", "admin@example.com", "noreply@example.com", "admin@example.com"},
		{"falls back to from_email", "", "noreply@example.com", "noreply@example.com"},
		{"falls back to default", "", "", "contato@uspavalia.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es := &EmailService{config: &config.Config{Email: config.Email{
				ContactEmail: tc.contactEmail,
				FromEmail:    tc.fromEmail,
			}}}
			if got := es.contactRecipient(); got != tc.want {
				t.Errorf("contactRecipient() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSendContactEmailUsesContactAddressAndReplyTo(t *testing.T) {
	cfg := &config.Config{Email: config.Email{
		FromEmail:    "noreply@example.com",
		FromName:     "Test",
		ContactEmail: "admin@example.com",
	}}
	es, fake := newTestEmailService(t, cfg)

	if err := es.SendContactEmail("Ana", "Silva", "ana@example.com", "Olá"); err != nil {
		t.Fatalf("SendContactEmail: %v", err)
	}
	if fake.input == nil {
		t.Fatal("SES SendEmail was not called")
	}
	if got := fake.input.Destination.ToAddresses; len(got) != 1 || got[0] != "admin@example.com" {
		t.Errorf("ToAddresses = %v, want [admin@example.com]", got)
	}
	if got := fake.input.ReplyToAddresses; len(got) != 1 || got[0] != "ana@example.com" {
		t.Errorf("ReplyToAddresses = %v, want [ana@example.com]", got)
	}
	if got := *fake.input.FromEmailAddress; got != "Test <noreply@example.com>" {
		t.Errorf("FromEmailAddress = %q", got)
	}
}

func TestSendMagicLinkHasNoReplyTo(t *testing.T) {
	es, fake := newTestEmailService(t, &config.Config{Email: config.Email{FromEmail: "noreply@example.com"}})

	if err := es.SendMagicLink("user@example.com", "https://example.com/login?t=1"); err != nil {
		t.Fatalf("SendMagicLink: %v", err)
	}
	if got := fake.input.Destination.ToAddresses; len(got) != 1 || got[0] != "user@example.com" {
		t.Errorf("ToAddresses = %v, want [user@example.com]", got)
	}
	if len(fake.input.ReplyToAddresses) != 0 {
		t.Errorf("ReplyToAddresses = %v, want none", fake.input.ReplyToAddresses)
	}
}

func TestSendEmailWithoutClient(t *testing.T) {
	es := &EmailService{config: &config.Config{}}
	if err := es.SendEmail("a@example.com", "", EmailTemplate{}); err == nil {
		t.Error("expected error when SES client is not configured")
	}
}
