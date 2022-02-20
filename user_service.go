package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (a *app) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	return a.RepositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.UserRepository.ConfirmUser(ctx, userID)
		},
	)
}

func (a *app) SetUpNewUser(ctx context.Context, user *User) error {
	// Create Organization
	organization := &Organization{
		ID:    uuid.New(),
		Title: user.Name,
	}

	// Create User
	user.ID = uuid.New()
	user.OrganizationID = organization.ID

	confirmationID := uuid.New()

	// Create initial project
	project := &Project{
		ID:             uuid.New(),
		Title:          "My Project",
		Active:         true,
		OrganizationID: organization.ID,
	}

	// Send email confirmation link
	subject := "Confirm your Email Address"
	body := fmt.Sprintf(
		`Confirm your e-mail address at %v/signup/confirm/%v to activate your account.`,
		a.Config.Webroot,
		confirmationID,
	)

	return a.RepositoryTxer.InTx(
		ctx,
		// Create Organization
		func(ctx context.Context) error {
			_, err := a.OrganizationRepository.InsertOrganization(ctx, organization)
			if err != nil {
				return err
			}
			return nil
		},
		// Create User
		func(ctx context.Context) error {
			_, err := a.UserRepository.InsertUserWithConfirmationID(ctx, user, confirmationID)
			if err != nil {
				return err
			}
			return nil
		},
		// Create initial project
		func(ctx context.Context) error {
			_, err := a.ProjectRepository.InsertProject(ctx, project)
			if err != nil {
				return err
			}
			return nil
		},
		// Send email confirmation link
		func(ctx context.Context) error {
			return a.MailResource.SendMail(user.EMail, subject, body)
		},
	)
}
