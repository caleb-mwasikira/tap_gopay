package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	db "github.com/caleb-mwasikira/tap_gopay/database"
	"github.com/caleb-mwasikira/tap_gopay/utils"
	"gopkg.in/gomail.v2"
)

func sendEmail(email, subject string, body []byte) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	authEmail := os.Getenv("AUTH_EMAIL")
	authPass := os.Getenv("AUTH_PASSWORD")

	// create new email msg
	m := gomail.NewMessage()
	m.SetHeader("From", authEmail)
	m.SetHeader("Subject", subject)
	m.SetHeader("To", email)
	m.SetBody("text/html", string(body))

	// send email
	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		log.Fatalf("invalid SMTP_PORT environment variable")
	}

	d := gomail.NewDialer(smtpHost, int(port), authEmail, authPass)
	return d.DialAndSend(m)
}

func sendOtpEmail(email, name, otp string) error {
	tmplFile := filepath.Join(utils.EmailViewsDir, "otp_email.html")
	t, err := template.ParseFiles(tmplFile)
	if err != nil {
		return err
	}

	tmplData := struct {
		Name        string
		Otp         []string
		CurrentYear int
	}{
		Name:        name,
		Otp:         utils.StringToRuneSlice(otp),
		CurrentYear: time.Now().Year(),
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, tmplData)
	if err != nil {
		return err
	}

	err = sendEmail(email, "TapGoPay Email Verification", buff.Bytes())
	return err
}

func sendWelcomeEmail(email string) error {
	dbUser, err := db.GetUser(email)
	if err != nil {
		return err
	}

	tmplData := struct {
		Name        string
		CurrentYear string
	}{
		Name:        dbUser.Username,
		CurrentYear: fmt.Sprintln(time.Now().Year()),
	}

	tmplPath := filepath.Join(utils.EmailViewsDir, "welcome_email.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, tmplData)
	if err != nil {
		return err
	}

	err = sendEmail(email, "Welcome To TapGoPay", buff.Bytes())
	return err
}
