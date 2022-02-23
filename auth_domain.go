package main

import "github.com/google/uuid"

type contextKey int

const contextKeyPrincipal contextKey = 0

type Principal struct {
	Name           string
	Username       string
	OrganizationID uuid.UUID
	Roles          []string
}

func (p *Principal) HasRole(role string) bool {
	for _, c := range p.Roles {
		if c == role {
			return true
		}
	}
	return false
}
