package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (a *app) SetUpNewUser(ctx context.Context, user *User) error {
	// Create user and organization
	user.ID = uuid.New()
	user.OrganizationID = uuid.New()

	confirmationID := uuid.New()

	_, err := a.UserRepository.InsertUserWithOrganizationAndConfirmation(ctx, user, confirmationID)
	if err != nil {
		return err
	}

	// Create initial project
	project := &Project{
		ID:             uuid.New(),
		Title:          "My Project",
		Active:         true,
		OrganizationID: user.OrganizationID,
	}
	_, err = a.ProjectRepository.InsertProject(ctx, project)
	if err != nil {
		return err
	}

	// Send Email confirmation link
	subject := "Confirm your Email Address"
	body := fmt.Sprintf(
		`Confirm your e-mail address at %v/signup/confirm/%v to activate your account.`,
		a.Config.Webroot,
		confirmationID,
	)

	return a.MailService.SendMail(user.EMail, subject, body)
}
