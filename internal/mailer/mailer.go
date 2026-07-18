package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"go-web-template/internal/config"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))

type MailerInterface interface {
	SendConfirmationEmail(to, name, link string) error
	SendPasswordResetEmail(to, name, link string) error
}

type Mailer struct {
	dialer *gomail.Dialer
	cfg    config.MailerConfig
	logger *zap.Logger
}

var _ MailerInterface = (*Mailer)(nil)

// NewMailer runs in console mode when cfg.Host is empty: Send* logs the link
// instead of dialing SMTP, so the template runs with zero mail setup.
func NewMailer(cfg config.MailerConfig, logger *zap.Logger) *Mailer {
	var dialer *gomail.Dialer
	if cfg.Host != "" {
		dialer = gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	}
	return &Mailer{dialer: dialer, cfg: cfg, logger: logger}
}

func (m *Mailer) SendConfirmationEmail(to, name, link string) error {
	return m.send(to, "confirm.html", "Please confirm your email address", "confirmation", name, link)
}

func (m *Mailer) SendPasswordResetEmail(to, name, link string) error {
	return m.send(to, "reset.html", "Reset your password", "password reset", name, link)
}

func (m *Mailer) send(to, tmpl, subject, kind, name, link string) error {
	if m.dialer == nil {
		m.logger.Info("mailer in console mode; email not sent",
			zap.String("kind", kind),
			zap.String("to", to),
			zap.String("link", link),
		)
		return nil
	}

	var body bytes.Buffer
	if err := templates.ExecuteTemplate(&body, tmpl, map[string]string{"Name": name, "Link": link}); err != nil {
		return fmt.Errorf("failed to render %s email: %w", kind, err)
	}

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.cfg.From, m.cfg.FromName)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	return m.dialer.DialAndSend(msg)
}
