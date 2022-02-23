package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type MailResource interface {
	SendMail(to, subject, body string) error
}

// SmtpMailResource is a SMTP based mail service
type SmtpMailResource struct {
	SMTPServername string
	SMTPFrom       string
	SMTPUser       string
	SMTPPassword   string
}

var _ MailResource = (*SmtpMailResource)(nil)

// NewSmtpMailResource creates a new SMTP based mail service
func NewSmtpMailResource(
	SMTPServername string,
	SMTPFrom string,
	SMTPUser string,
	SMTPPassword string) *SmtpMailResource {
	return &SmtpMailResource{
		SMTPServername: SMTPServername,
		SMTPFrom:       SMTPFrom,
		SMTPUser:       SMTPUser,
		SMTPPassword:   SMTPPassword,
	}
}

func (s *SmtpMailResource) SendMail(to, subject, body string) error {
	fromAddress := mail.Address{
		Name:    "Baralga Time Tracker",
		Address: s.SMTPFrom,
	}
	toAddress := mail.Address{
		Name:    "",
		Address: to,
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = fromAddress.String()
	headers["To"] = toAddress.String()
	headers["Subject"] = subject

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := s.SMTPServername

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", s.SMTPUser, s.SMTPPassword, host)

	useTls := strings.Contains(servername, "465")
	var client *smtp.Client
	if useTls {
		// TLS config
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}

		// Here is the key, you need to call tls.Dial instead of smtp.Dial
		// for smtp servers running on 465 that require an ssl connection
		// from the very beginning (no starttls)
		conn, err := tls.Dial("tcp", servername, tlsconfig)
		if err != nil {
			log.Panic(err)
		}

		c, err := smtp.NewClient(conn, host)
		if err != nil {
			log.Panic(err)
		}

		client = c
	} else {
		c, err := smtp.Dial(servername)
		if err != nil {
			log.Panic(err)
		}

		client = c
	}

	defer func() {
		_ = client.Quit()
	}()

	// Auth
	if err := client.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err := client.Mail(fromAddress.Address); err != nil {
		return err
	}

	if err := client.Rcpt(toAddress.Address); err != nil {
		return err
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}
