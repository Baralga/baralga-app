package shared

import (
	"fmt"
)

type InMemMailResource struct {
	Mails []string
}

var _ MailResource = (*InMemMailResource)(nil)

func NewInMemMailResource() *InMemMailResource {
	return &InMemMailResource{
		Mails: make([]string, 0),
	}
}

func (s *InMemMailResource) SendMail(to, subject, body string) error {
	mail := fmt.Sprintf("%v\n %v\n %v\n", to, subject, body)
	s.Mails = append(s.Mails, mail)
	return nil
}
