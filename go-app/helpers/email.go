package helpers

import (
	"bytes"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
)

func GetTemplatePath(templateName string) string {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	path := filepath.Join(dir, "templates", templateName)

	return path
}

type EmailData struct {
	Name    string
	Email   string
	Link    string
	Subject string
}

func SendEmail(to []string, templateName string, data EmailData) (bool, string) {
	templatePath := GetTemplatePath(templateName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return false, err.Error()
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return false, err.Error()
	}
	smtpHost := os.Getenv("GMAIL_HOST")
	smtpPort := os.Getenv("GMAIL_PORT")
	smtpUser := os.Getenv("FROM_EMAIL")
	smtpPass := os.Getenv("GOOGLE_EMAIL_PASSWORD")

	from := smtpUser
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	msg := "Subject: " + data.Subject + "\n" + headers + "\n\n" + body.String()

	err = smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		to,
		[]byte(msg),
	)
	if err != nil {
		log.Printf("error sending email: %v", err)
		return false, err.Error()
	}
	return true, ""
}
