package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InMemOrganizationInviteRepository struct {
	invites []*OrganizationInvite
}

var _ OrganizationInviteRepository = (*InMemOrganizationInviteRepository)(nil)

func NewInMemOrganizationInviteRepository() *InMemOrganizationInviteRepository {
	return &InMemOrganizationInviteRepository{
		invites: []*OrganizationInvite{},
	}
}

func (r *InMemOrganizationInviteRepository) InsertInvite(ctx context.Context, invite *OrganizationInvite) (*OrganizationInvite, error) {
	r.invites = append(r.invites, invite)
	return invite, nil
}

func (r *InMemOrganizationInviteRepository) FindInviteByToken(ctx context.Context, token string) (*OrganizationInvite, error) {
	for _, invite := range r.invites {
		if invite.Token == token {
			return invite, nil
		}
	}
	return nil, ErrInviteNotFound
}

func (r *InMemOrganizationInviteRepository) FindInvitesByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*OrganizationInvite, error) {
	var result []*OrganizationInvite
	for _, invite := range r.invites {
		if invite.OrgID == organizationID {
			result = append(result, invite)
		}
	}
	return result, nil
}

func (r *InMemOrganizationInviteRepository) UpdateInvite(ctx context.Context, invite *OrganizationInvite) error {
	for i, existingInvite := range r.invites {
		if existingInvite.ID == invite.ID {
			r.invites[i] = invite
			return nil
		}
	}
	return errors.New("invite not found")
}

func (r *InMemOrganizationInviteRepository) DeleteInvite(ctx context.Context, inviteID uuid.UUID) error {
	for i, invite := range r.invites {
		if invite.ID == inviteID {
			r.invites = append(r.invites[:i], r.invites[i+1:]...)
			return nil
		}
	}
	return errors.New("invite not found")
}
