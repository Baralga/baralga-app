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

		// Get organization details
		organization, err := h.organizationService.GetOrganization(r.Context(), orgID)
		if err != nil {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		// Set HX-Trigger to show modal
		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		// Render organization management dialog
		h.renderOrganizationDialog(w, r, organization, principal)
	}
}

// HandleOrganizationTitleUpdate handles organization title updates
func (h *OrganizationWebHandlers) HandleOrganizationTitleUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user and organization from context
		principal := shared.MustPrincipalFromContext(r.Context())
		orgID := principal.OrganizationID

		// Check if user is admin using Principal roles
		if !principal.HasRole("ROLE_ADMIN") {
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

		if len(title) > 100 {
			h.renderValidationError(w, r, "Organization name must be between 1 and 100 characters")
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

		w.Header().Set("HX-Trigger", "{ \"baralga__main_content_modal-hide\": true }")

		// Render success response
		h.renderSuccess(w, r, "Organization name updated successfully.")
	}
}

// renderOrganizationDialog renders the organization management dialog
func (h *OrganizationWebHandlers) renderOrganizationDialog(w http.ResponseWriter, r *http.Request, organization *Organization, principal *shared.Principal) {
	shared.RenderHTML(w, h.OrganizationDialog(organization, principal))
}

// OrganizationDialog returns the organization management dialog as a gomponents node
func (h *OrganizationWebHandlers) OrganizationDialog(organization *Organization, principal *shared.Principal) g.Node {
	// Check if user is admin
	isAdmin := principal.HasRole("ROLE_ADMIN")

	// Create input attributes based on admin status
	inputAttrs := []g.Node{
		Type("text"),
		Class("form-control"),
		ID("orgTitle"),
		Name("title"),
		Value(organization.Title),
		MaxLength("100"),
	}

	// Add readonly attribute for non-admin users
	if !isAdmin {
		inputAttrs = append(inputAttrs, ReadOnly())
	} else {
		inputAttrs = append(inputAttrs, Required())
	}

	// Create form attributes based on admin status
	formAttrs := []g.Node{
		ID("baralga__main_content_modal_content"),
		Class("modal-content"),
	}

	// Add form submission attributes only for admin users
	if isAdmin {
		formAttrs = append(formAttrs,
			ghx.Post("/profile/organization"),
			ghx.Target("#baralga__main_content_modal_content"),
			ghx.Swap("outerHTML"),
		)
	}

	// Build all form content
	formContent := []g.Node{
		Div(
			Class("modal-header"),
			H2(
				Class("modal-title"),
				g.Text("Organization Settings"),
			),
			A(
				g.Attr("data-bs-dismiss", "modal"),
				Class("btn-close"),
			),
		),
		Div(
			Class("modal-body"),
			Div(
				Class("mb-3"),
				Label(
					For("orgTitle"),
					Class("form-label"),
					g.Text("Organization Name"),
				),
				Input(inputAttrs...),
			),
		),
		Div(
			Class("modal-footer"),
			// Only show save button for admin users
			g.If(isAdmin,
				Button(
					Type("submit"),
					Class("btn btn-primary"),
					I(Class("bi-save me-2")),
					g.Text("Save Changes"),
				),
			),
			A(
				g.Attr("data-bs-dismiss", "modal"),
				Class("btn btn-secondary"),
				I(Class("bi-x me-2")),
				g.Text("Cancel"),
			),
		),
	}

	// Combine form attributes with content
	allAttrs := append(formAttrs, formContent...)
	return FormEl(allAttrs...)
}

// renderSuccess renders a success message
func (h *OrganizationWebHandlers) renderSuccess(w http.ResponseWriter, r *http.Request, message string) {
	shared.RenderHTML(w, h.SuccessMessage(message))
}

// SuccessMessage returns a success message as a gomponents node
func (h *OrganizationWebHandlers) SuccessMessage(message string) g.Node {
	return Div(
		Class("alert alert-success text-center"),
		Role("alert"),
		Span(g.Text(message)),
	)
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
