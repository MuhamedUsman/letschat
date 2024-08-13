package mailer

import (
	"bytes"
	"embed"
	"github.com/M0hammadUsman/letschat/internal/common"
	"gopkg.in/gomail.v2"
	"html/template"
)

//go:embed templates
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender string
}

func New(cfg *common.Config) *Mailer {
	return &Mailer{
		dialer: gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password),
		sender: cfg.SMTP.Sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}
	html := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(html, "body", data); err != nil {
		return err
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("To", recipient)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/html", html.String())
	if err = m.dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
