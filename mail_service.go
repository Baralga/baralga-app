package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
)

type MailService interface {
	SendMail(to, subject, body string) error
}

// SmtpMailService is a SMTP based mail service
type SmtpMailService struct {
	SMTPServername string
	SMTPFrom       string
	SMTPUser       string
	SMTPPassword   string
}

var _ MailService = (*SmtpMailService)(nil)

// NewSmtpMailService creates a new SMTP based mail service
func NewSmtpMailService(
	SMTPServername string,
	SMTPFrom string,
	SMTPUser string,
	SMTPPassword string) *SmtpMailService {
	return &SmtpMailService{
		SMTPServername: SMTPServername,
		SMTPFrom:       SMTPFrom,
		SMTPUser:       SMTPUser,
		SMTPPassword:   SMTPPassword,
	}
}

func (s *SmtpMailService) SendMail(to, subject, body string) error {
	fromAddress := mail.Address{
		Name:    "Baralga Time Tracker",
		Address: s.SMTPFrom,
	}
	toAddress := mail.Address{
		Name:    "",
		Address: to,
	}
	//	subj := "TEST: This is the email subject"
	//	body := "This is an example body.\n With two lines."

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
	defer c.Quit()
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(fromAddress.Address); err != nil {
		return err
	}

	if err = c.Rcpt(toAddress.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
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
