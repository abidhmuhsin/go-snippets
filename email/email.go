package main

import (
	"fmt"
	"net/smtp"

	"github.com/go-errors/errors"
)

// SMTP_FROM=<smtp_from_email_address>
// SMTP_HOST=<smtp_host> smtp.gmail.com
// SMTP_PORT=<smtp_port>  465
// SMTP_USER=<smtp_user>
// SMTP_PASSWORD=<smtp_password>

const (
	// SMTP_FROM     = "abidhmuhsin@gmail.com"
	// SMTP_HOST     = "smtp.gmail.com"
	// SMTP_PORT     = "587" //"465"
	// SMTP_USER     = "abidhmuhsin@gmail.com"
	// SMTP_PASSWORD = "ymubyhnsmlhvvise"

	SMTP_FROM     = "abidhmuhsin@gmail.com"
	SMTP_HOST     = "localhost"
	SMTP_PORT     = "25"
	SMTP_USER     = ""
	SMTP_PASSWORD = ""
)

type SendInput struct {
	Template     TemplateName
	TemplateVars map[string]string
	To           string
}

func Send(input *SendInput) error {
	template, err := GetTemplate(input.Template, input.TemplateVars)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	from := SMTP_FROM

	headers := map[string]string{
		"Content-Type": "text/html; charset=utf-8",
		"From":         from,
		"MIME-Version": "1.0",
		"Subject":      template.Subject,
		"To":           input.To,
	}

	msg := []byte{}

	for k, v := range headers {
		msg = append(msg, []byte(fmt.Sprintf("%s: %s\n", k, v))...)
	}

	msg = append(msg, []byte("\n")...)
	msg = append(msg, template.Html...)

	var auth smtp.Auth

	addr := fmt.Sprintf("%s:%s", SMTP_HOST, SMTP_PORT)

	smtpUser := SMTP_USER
	smtpPassword := SMTP_PASSWORD

	if smtpPassword != "" && smtpUser != "" {
		auth = smtp.PlainAuth("", SMTP_USER, SMTP_PASSWORD, SMTP_HOST)
	}

	fmt.Println(addr, auth)
	return smtp.SendMail(addr, auth, from, []string{input.To}, msg)
}
