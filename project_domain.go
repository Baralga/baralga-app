package main

import "github.com/google/uuid"

type Project struct {
	ID             uuid.UUID
	Title          string
	Description    string
	Active         bool
	OrganizationID uuid.UUID
}
