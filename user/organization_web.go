package user

import (
	"net/http"
	"strings"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html" //nolint:all
)

// OrganizationWebHandlers handles organization web requests
type OrganizationWebHandlers struct {
	config              *shared.Config
	organizationService OrganizationService
}

// NewOrganizationWebHandlers creates new organization web handlers
func NewOrganizationWebHandlers(config *shared.Config, organizationService OrganizationService) *OrganizationWebHandlers {
	return &OrganizationWebHandlers{
		config:              config,
		organizationService: organizationService,
	}
}

// HandleOrganizationManagementPage returns the organization management dialog
func (h *OrganizationWebHandlers) HandleOrganizationManagementPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user and organization from context
		principal := shared.MustPrincipalFromContext(r.Context())
		orgID := principal.OrganizationID

		// Check if user is admin using Principal roles
		if !principal.HasRole("admin") {
			http.Error(w, "Access denied. Administrator privileges required.", http.StatusForbidden)
			return
		}

		// Get organization details
		organization, err := h.organizationService.GetOrganization(r.Context(), orgID)
		if err != nil {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		// Render organization management dialog
		h.renderOrganizationDialog(w, r, organization)
	}
}

// HandleOrganizationTitleUpdate handles organization title updates
func (h *OrganizationWebHandlers) HandleOrganizationTitleUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user and organization from context
		principal := shared.MustPrincipalFromContext(r.Context())
		orgID := principal.OrganizationID

		// Check if user is admin using Principal roles
		if !principal.HasRole("admin") {
			h.renderError(w, r, "Access denied. Administrator privileges required.", http.StatusForbidden)
			return
		}

		// Parse form data
		title := strings.TrimSpace(r.FormValue("title"))

		// Validate input
		if title == "" {
			h.renderValidationError(w, r, "Organization name is required")
			return
		}

		if len(title) > 255 {
			h.renderValidationError(w, r, "Organization name must be between 1 and 255 characters")
			return
		}

		// Update organization name
		err := h.organizationService.UpdateOrganizationName(r.Context(), orgID, title)
		if err != nil {
			if err != nil && err.Error() == "Organization name already exists" {
				h.renderValidationError(w, r, "Organization name already exists. Please choose a different name.")
				return
			}
			if err == shared.ErrNotFound {
				h.renderError(w, r, "Organization not found", http.StatusNotFound)
				return
			}
			h.renderError(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Render success response
		h.renderSuccess(w, r, "Organization name updated successfully.")
	}
}

// renderOrganizationDialog renders the organization management dialog
func (h *OrganizationWebHandlers) renderOrganizationDialog(w http.ResponseWriter, r *http.Request, organization *Organization) {
	shared.RenderHTML(w, h.OrganizationDialog(organization))
}

// OrganizationDialog returns the organization management dialog as a gomponents node
func (h *OrganizationWebHandlers) OrganizationDialog(organization *Organization) g.Node {
	return Div(
		ID("organizationModal"),
		Class("modal fade"),
		g.Attr("tabindex", "-1"),
		Div(
			Class("modal-dialog"),
			Div(
				Class("modal-content"),
				Div(
					Class("modal-header"),
					H5(
						Class("modal-title"),
						g.Text("Organization Settings"),
					),
					Button(
						Type("button"),
						Class("btn-close"),
						g.Attr("data-bs-dismiss", "modal"),
					),
				),
				Div(
					Class("modal-body"),
					Form(
						ghx.Post("/profile/organization"),
						ghx.Target("#organizationForm"),
						Div(
							Class("mb-3"),
							Label(
								For("orgTitle"),
								Class("form-label"),
								g.Text("Organization Name"),
							),
							Input(
								Type("text"),
								Class("form-control"),
								ID("orgTitle"),
								Name("title"),
								Value(organization.Title),
								Required(),
							),
						),
						Div(
							Class("modal-footer"),
							Button(
								Type("button"),
								Class("btn btn-secondary"),
								g.Attr("data-bs-dismiss", "modal"),
								g.Text("Cancel"),
							),
							Button(
								Type("submit"),
								Class("btn btn-primary"),
								g.Text("Save Changes"),
							),
						),
					),
				),
			),
		),
	)
}

// renderSuccess renders a success message
func (h *OrganizationWebHandlers) renderSuccess(w http.ResponseWriter, r *http.Request, message string) {
	shared.RenderHTML(w, h.SuccessMessage(message))
}

// SuccessMessage returns a success message as a gomponents node
func (h *OrganizationWebHandlers) SuccessMessage(message string) g.Node {
	return g.Group([]g.Node{
		Div(
			Class("alert alert-success"),
			g.Text(message),
		),
		Script(
			g.Text(`
				// Close modal and refresh page
				bootstrap.Modal.getInstance(document.getElementById('organizationModal')).hide();
				location.reload();
			`),
		),
	})
}

// renderValidationError renders validation error messages
func (h *OrganizationWebHandlers) renderValidationError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusBadRequest)
	shared.RenderHTML(w, h.ValidationErrorMessage(message))
}

// ValidationErrorMessage returns a validation error message as a gomponents node
func (h *OrganizationWebHandlers) ValidationErrorMessage(message string) g.Node {
	return Div(
		Class("alert alert-danger"),
		Ul(
			Li(g.Text(message)),
		),
	)
}

// renderError renders an error message
func (h *OrganizationWebHandlers) renderError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	w.WriteHeader(statusCode)
	shared.RenderHTML(w, h.ErrorMessage(message))
}

// ErrorMessage returns an error message as a gomponents node
func (h *OrganizationWebHandlers) ErrorMessage(message string) g.Node {
	return Div(
		Class("alert alert-danger"),
		g.Text(message),
	)
}

// RegisterProtected registers protected routes (requires authentication)
func (h *OrganizationWebHandlers) RegisterProtected(router chi.Router) {
	router.Get("/profile/organization", h.HandleOrganizationManagementPage())
	router.Post("/profile/organization", h.HandleOrganizationTitleUpdate())
}

// RegisterOpen registers open routes (no authentication required)
func (h *OrganizationWebHandlers) RegisterOpen(router chi.Router) {
	// No open routes for organization management
}
