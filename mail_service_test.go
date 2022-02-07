package main

import "fmt"

type InMemMailService struct {
	mails []string
}

var _ MailService = (*InMemMailService)(nil)

func NewInMemMailService() *InMemMailService {
	return &InMemMailService{
		mails: make([]string, 0),
	}
}

func (s *InMemMailService) SendMail(to, subject, body string) error {
	mail := fmt.Sprintf("%v\n %v\n %v\n", to, subject, body)
	s.mails = append(s.mails, mail)
	return nil
}
