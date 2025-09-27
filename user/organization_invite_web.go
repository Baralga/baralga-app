package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html" //nolint:all
)

type OrganizationInviteWebHandlers struct {
	config      *shared.Config
	userService *UserService
}

func NewOrganizationInviteWeb(config *shared.Config, userService *UserService) *OrganizationInviteWebHandlers {
	return &OrganizationInviteWebHandlers{
		config:      config,
		userService: userService,
	}
}

func (h *OrganizationInviteWebHandlers) RegisterProtected(r chi.Router) {
	r.Get("/organization/invites", h.HandleInviteList())
	r.Post("/organization/invites/generate", h.HandleGenerateInvite())
}

func (h *OrganizationInviteWebHandlers) HandleInviteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		// Get all invites for the organization (only admins can see invites)
		var invites []*OrganizationInvite
		if principal.HasRole("ROLE_ADMIN") {
			var err error
			invites, err = h.userService.FindOrganizationInvites(r.Context(), principal)
			if err != nil {
				http.Error(w, "Failed to load invites", http.StatusInternalServerError)
				return
			}
		}

		shared.RenderHTML(w, h.InviteList(principal, invites))
	}
}

func (h *OrganizationInviteWebHandlers) HandleGenerateInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		// Only admins can generate invites
		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Generate new invite
		_, err := h.userService.GenerateOrganizationInvite(r.Context(), principal)
		if err != nil {
			http.Error(w, "Failed to generate invite", http.StatusInternalServerError)
			return
		}

		// Return the invite list with the new invite
		invites, err := h.userService.FindOrganizationInvites(r.Context(), principal)
		if err != nil {
			http.Error(w, "Failed to load invites", http.StatusInternalServerError)
			return
		}

		shared.RenderHTML(w, h.InviteList(principal, invites))
	}
}

func (h *OrganizationInviteWebHandlers) InviteList(principal *shared.Principal, invites []*OrganizationInvite) g.Node {
	return Div(
		Class("container-fluid"),
		Div(
			Class("d-flex justify-content-between align-items-center mb-3"),
			H3(
				Class("mb-0"),
				I(Class("bi-people me-2")),
				g.Text("Organization Invites"),
			),
			g.If(principal.HasRole("ROLE_ADMIN"),
				Button(
					Type("button"),
					Class("btn btn-primary"),
					ghx.Post("/organization/invites/generate"),
					ghx.Target("#invite-list"),
					ghx.Swap("outerHTML"),
					I(Class("bi-plus me-2")),
					g.Text("Generate Invite"),
				),
			),
		),
		Div(
			ID("invite-list"),
			g.If(len(invites) == 0,
				Div(
					Class("alert alert-info"),
					I(Class("bi-info-circle me-2")),
					g.Text("No invites have been generated yet."),
				),
			),
			g.If(len(invites) > 0,
				func() g.Node {
					var nodes []g.Node
					for _, invite := range invites {
						nodes = append(nodes, h.InviteCard(invite))
					}
					return Div(append([]g.Node{Class("row")}, nodes...)...)
				}(),
			),
		),
	)
}

func (h *OrganizationInviteWebHandlers) InviteCard(invite *OrganizationInvite) g.Node {
	status := h.getInviteStatus(invite)
	statusClass := h.getStatusClass(status)

	return Div(
		Class("col-md-6 col-lg-4 mb-3"),
		Div(
			Class("card h-100"),
			Div(
				Class("card-header d-flex justify-content-between align-items-center"),
				Span(
					Class("badge "+statusClass),
					g.Text(status),
				),
				Small(
					Class("text-muted"),
					g.Text(invite.CreatedAt.Format("Jan 2, 15:04")),
				),
			),
			Div(
				Class("card-body"),
				Div(
					Class("mb-2"),
					Strong(g.Text("Invite Link:")),
				),
				Div(
					Class("input-group"),
					Input(
						Type("text"),
						Class("form-control font-monospace"),
						Value(h.getInviteURL(invite.Token)),
						g.Attr("readonly", "readonly"),
						ID(fmt.Sprintf("invite-url-%s", invite.ID.String())),
					),
					Button(
						Type("button"),
						Class("btn btn-outline-secondary"),
						g.Attr("onclick", fmt.Sprintf("copyToClipboard('invite-url-%s')", invite.ID.String())),
						I(Class("bi-copy")),
					),
				),
				func() g.Node {
					if invite.UsedAt != nil {
						return Div(
							Class("mt-2"),
							Small(
								Class("text-muted"),
								g.Textf("Used on %s", invite.UsedAt.Format("Jan 2, 15:04")),
							),
						)
					}
					return g.Text("")
				}(),
				g.If(invite.ExpiresAt.Before(time.Now()) && invite.UsedAt == nil,
					Div(
						Class("mt-2"),
						Small(
							Class("text-danger"),
							g.Textf("Expired on %s", invite.ExpiresAt.Format("Jan 2, 15:04")),
						),
					),
				),
			),
		),
	)
}

func (h *OrganizationInviteWebHandlers) getInviteStatus(invite *OrganizationInvite) string {
	if invite.UsedAt != nil {
		return "Used"
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return "Expired"
	}
	return "Active"
}

func (h *OrganizationInviteWebHandlers) getStatusClass(status string) string {
	switch status {
	case "Used":
		return "bg-success"
	case "Expired":
		return "bg-danger"
	case "Active":
		return "bg-primary"
	default:
		return "bg-secondary"
	}
}

func (h *OrganizationInviteWebHandlers) getInviteURL(token string) string {
	return fmt.Sprintf("%s/signup?invite=%s", h.config.Webroot, token)
}
