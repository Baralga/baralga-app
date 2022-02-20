package main

import "fmt"

type InMemMailResource struct {
	mails []string
}

var _ MailResource = (*InMemMailResource)(nil)

func NewInMemMailResource() *InMemMailResource {
	return &InMemMailResource{
		mails: make([]string, 0),
	}
}

func (s *InMemMailResource) SendMail(to, subject, body string) error {
	mail := fmt.Sprintf("%v\n %v\n %v\n", to, subject, body)
	s.mails = append(s.mails, mail)
	return nil
}
