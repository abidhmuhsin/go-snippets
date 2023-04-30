package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/go-errors/errors"
)

type RawTemplate *struct {
	Html    *template.Template
	Subject string
}

type Template struct {
	Html    []byte
	Subject string
}

type TemplateName string

const (
	basePath = "templates"

	TemplateWelcome          TemplateName = "welcome"
	TemplatePasswordRecovery TemplateName = "password_reset"
)

var (
	templates = map[TemplateName]RawTemplate{
		TemplatePasswordRecovery: {
			Subject: "Password reset request",
		},
		TemplateWelcome: {
			Subject: "Welcome to The Dev Company",
		},
	}
)

func GetTemplate(name TemplateName, data map[string]string) (*Template, error) {
	var err error

	t := templates[name]

	if t.Html == nil {
		templatePath := filepath.Join(basePath, fmt.Sprintf("%s.gohtml", name))

		t.Html, err = template.ParseFiles(templatePath)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}

	buffer := &bytes.Buffer{}

	err = t.Html.Execute(buffer, data)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	return &Template{
		Html:    buffer.Bytes(),
		Subject: strings.ReplaceAll(t.Subject, "[OrganizationName]", data["OrganizationName"]),
	}, nil
}
