package user

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html" //nolint:all
)

type signupFormModel struct {
	CSRFToken        string
	Name             string `validate:"required,min=5,max=50"`
	EMail            string `validate:"required,email"`
	Password         string `validate:"required,min=8,max=100"`
	AcceptConditions bool
	InviteToken      string
}

type organizationFormModel struct {
	CSRFToken string
	Name      string `validate:"required,min=1,max=255"`
}

type UserWebHandlers struct {
	config         *shared.Config
	userService    *UserService
	userRepository UserRepository
}

func NewUserWeb(config *shared.Config, userService *UserService, userRepository UserRepository) *UserWebHandlers {
	return &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}
}

func (a *UserWebHandlers) RegisterProtected(r chi.Router) {
	r.Get("/organization/dialog", a.HandleOrganizationDialog())
	r.Post("/organization/update", a.HandleOrganizationUpdate())
	r.Get("/organization/invites", a.HandleOrganizationInvites())
	r.Post("/organization/invites/generate", a.HandleGenerateInviteInDialog())
	r.Get("/signup/invite/{token}", a.HandleInviteSignUpPage())
}

func (a *UserWebHandlers) RegisterOpen(r chi.Router) {
	r.Get("/signup", a.HandleSignUpPage())
	r.Post("/signup", a.HandleSignUpForm())
	r.Post("/signup/validate", a.HandleSignUpFormValidate())
	r.Get("/signup/confirm/{confirmation-id}", a.HandleSignUpConfirm())
}

func (a *UserWebHandlers) signupFormValidator(incomplete bool) func(ctx context.Context, formModel signupFormModel) (map[string]string, error) {
	validator := validator.New()
	userRepository := a.userRepository
	return func(ctx context.Context, formModel signupFormModel) (map[string]string, error) {
		if !incomplete {
			err := validator.Struct(formModel)
			if err != nil {
				return nil, err
			}
		}

		fieldErrors := make(map[string]string)

		if formModel.EMail != "" {
			errs := validator.Var(formModel.EMail, "email")
			if errs != nil {
				fieldErrors["EMail"] = "Invalid email."
			}

			_, err := userRepository.FindUserByUsername(ctx, formModel.EMail)
			if !errors.Is(err, ErrUserNotFound) {
				fieldErrors["EMail"] = "Email not available."
			}
		}

		if len(fieldErrors) > 0 {
			return fieldErrors, errors.New("validation failed")
		}

		return fieldErrors, nil
	}
}

func (a *UserWebHandlers) HandleOrganizationDialog() http.HandlerFunc {
	userService := a.userService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		// Get the actual organization name
		organization, err := userService.FindOrganizationByID(r.Context(), principal.OrganizationID)
		if err != nil {
			// Fallback to a default name if organization not found
			organization = &Organization{
				ID:    principal.OrganizationID,
				Title: "Organization",
			}
		}

		formModel := organizationFormModel{
			CSRFToken: csrf.Token(r),
			Name:      organization.Title,
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		shared.RenderHTML(w, OrganizationDialog(principal, formModel))
	}
}

func (a *UserWebHandlers) HandleOrganizationUpdate() http.HandlerFunc {
	userService := a.userService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		err := r.ParseForm()
		if err != nil {
			formModel := organizationFormModel{
				CSRFToken: csrf.Token(r),
			}
			shared.RenderHTML(w, OrganizationDialog(principal, formModel))
			return
		}

		var formModel organizationFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, OrganizationDialog(principal, formModel))
			return
		}

		// Validate form
		validator := validator.New()
		err = validator.Struct(formModel)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, OrganizationDialog(principal, formModel))
			return
		}

		// Update organization name
		err = userService.UpdateOrganizationName(r.Context(), principal, formModel.Name)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, OrganizationDialog(principal, formModel))
			return
		}

		// Success - close modal and refresh page
		w.Header().Set("HX-Trigger", "baralga__main_content_modal-hide")
		shared.RenderHTML(w, Div())
	}
}

func (a *UserWebHandlers) HandleOrganizationInvites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		// Only admins can view invites
		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Get all invites for the organization
		invites, err := a.userService.FindOrganizationInvites(r.Context(), principal)
		if err != nil {
			http.Error(w, "Failed to load invites", http.StatusInternalServerError)
			return
		}

		shared.RenderHTML(w, OrganizationInvitesPage(a.config, principal, invites))
	}
}

func (a *UserWebHandlers) HandleGenerateInviteInDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		// Only admins can generate invites
		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Generate new invite
		invite, err := a.userService.GenerateOrganizationInvite(r.Context(), principal)
		if err != nil {
			http.Error(w, "Failed to generate invite", http.StatusInternalServerError)
			return
		}

		// Return the invite link display
		shared.RenderHTML(w, InviteLinkDisplay(a.config, invite))
	}
}

func (a *UserWebHandlers) HandleSignUpConfirm() http.HandlerFunc {
	userService := a.userService
	userRepository := a.userRepository
	return func(w http.ResponseWriter, r *http.Request) {
		confirmationIDParam := chi.URLParam(r, "confirmation-id")

		userID, err := userRepository.FindUserIDByConfirmationID(r.Context(), confirmationIDParam)
		if errors.Is(err, ErrUserNotFound) {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}
		if err != nil {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		err = userService.ConfirmUser(r.Context(), userID)
		if err != nil {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/login?info=confirm_successfull", http.StatusFound)
	}
}

func (a *UserWebHandlers) HandleSignUpFormValidate() http.HandlerFunc {
	validate := a.signupFormValidator(true)
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := signupFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", nil))
			return
		}

		var formModel signupFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", nil))
			return
		}

		fieldErrors, err := validate(r.Context(), formModel)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", fieldErrors))
			return
		}

		formModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, a.SignupForm(formModel, "", fieldErrors))
	}
}

func (a *UserWebHandlers) HandleSignUpPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		formModel := signupFormModel{}
		formModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, a.SignUpPage(r.URL.Path, formModel))
	}
}

func (a *UserWebHandlers) HandleSignUpForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validate := a.signupFormValidator(false)
	userService := a.userService
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := signupFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", nil))
			return
		}

		var formModel signupFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", nil))
			return
		}

		fieldErrors, err := validate(r.Context(), formModel)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.SignupForm(formModel, "", fieldErrors))
			return
		}

		user := mapSignUpFormToUser(formModel, userService.EncryptPassword(formModel.Password))

		// Check if this is an invite-based registration
		if formModel.InviteToken != "" {
			// Use invite-based registration
			err = userService.SetUpNewUserWithInvite(r.Context(), &user, formModel.InviteToken)
			if err != nil {
				shared.RenderProblemHTML(w, isProduction, err)
				return
			}
		} else {
			// Use regular registration (creates new organization)
			confirmationID := uuid.New()
			err = userService.SetUpNewUser(r.Context(), &user, confirmationID)
			if err != nil {
				shared.RenderProblemHTML(w, isProduction, err)
				return
			}
		}

		shared.RenderHTML(w, SignupSuccess(formModel))
	}
}

func (a *UserWebHandlers) HandleInviteSignUpPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		// Check if user is already logged in
		principal := a.getPrincipalFromContext(r.Context())
		if principal != nil {
			// User is logged in, show logout message
			shared.RenderHTML(w, a.InviteLogoutRequiredPage(principal.Username))
			return
		}

		// Validate the invite token and get invite details
		invite, err := a.userService.ValidateInvite(r.Context(), token)
		if err != nil {
			// Show error page for invalid/expired invite
			shared.RenderHTML(w, a.InviteErrorPage("Invalid or expired invite link"))
			return
		}

		// Fetch organization details
		organization, err := a.userService.FindOrganizationByID(r.Context(), invite.OrgID)
		if err != nil {
			shared.RenderHTML(w, a.InviteErrorPage("Unable to load organization details"))
			return
		}

		formModel := signupFormModel{
			CSRFToken:   csrf.Token(r),
			InviteToken: token,
		}

		shared.RenderHTML(w, a.InviteSignUpPage(r.URL.Path, formModel, organization))
	}
}

func (a *UserWebHandlers) SignUpPage(currentPath string, formModel signupFormModel) g.Node {
	return shared.Page(
		"Sign Up",
		currentPath,
		[]g.Node{
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("d-flex justify-content-center align-items-center mt-2 mb-3"),
						Img(
							Alt("Baralga"),
							Class("img-responsive"),
							Src("/assets/baralga_192.png"),
						),
						Div(
							Class("ms-4"),
							H2(
								g.Text("Baralga"),
								Small(
									Class("text-muted"),
									StyleAttr("display: block; font-size: 70%;"),
									g.Text("project time tracking"),
								),
							),
						),
					),
					a.SignupForm(formModel, "", nil),
				),
			),
		},
	)
}

func (a *UserWebHandlers) InviteSignUpPage(currentPath string, formModel signupFormModel, organization *Organization) g.Node {
	return shared.Page(
		"Join Organization",
		currentPath,
		[]g.Node{
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("d-flex justify-content-center align-items-center mt-2 mb-3"),
						Img(
							Alt("Baralga"),
							Class("img-responsive"),
							Src("/assets/baralga_192.png"),
						),
						Div(
							Class("ms-4"),
							H2(
								g.Text("Baralga"),
								Small(
									Class("text-muted"),
									StyleAttr("display: block; font-size: 70%;"),
									g.Text("project time tracking"),
								),
							),
						),
					),
					Div(
						Class("alert alert-info"),
						I(Class("bi-info-circle me-2")),
						g.Textf("You've been invited to join '%s'. Complete your registration below.", organization.Title),
					),
					a.InviteSignupForm(formModel, "", nil, organization),
				),
			),
		},
	)
}

func (a *UserWebHandlers) InviteErrorPage(message string) g.Node {
	return shared.Page(
		"Invalid Invite",
		"/signup/invite",
		[]g.Node{
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("d-flex justify-content-center align-items-center mt-2 mb-3"),
						Img(
							Alt("Baralga"),
							Class("img-responsive"),
							Src("/assets/baralga_192.png"),
						),
						Div(
							Class("ms-4"),
							H2(
								g.Text("Baralga"),
								Small(
									Class("text-muted"),
									StyleAttr("display: block; font-size: 70%;"),
									g.Text("project time tracking"),
								),
							),
						),
					),
					Div(
						Class("alert alert-danger"),
						I(Class("bi-exclamation-triangle me-2")),
						g.Text(message),
					),
					Div(
						Class("text-center"),
						A(
							Href("/signup"),
							Class("btn btn-primary"),
							g.Text("Create New Account"),
						),
					),
				),
			),
		},
	)
}

func (a *UserWebHandlers) InviteLogoutRequiredPage(username string) g.Node {
	return shared.Page(
		"Logout Required",
		"/signup/invite",
		[]g.Node{
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("d-flex justify-content-center align-items-center mt-2 mb-3"),
						Img(
							Alt("Baralga"),
							Class("img-responsive"),
							Src("/assets/baralga_192.png"),
						),
						Div(
							Class("ms-4"),
							H2(
								g.Text("Baralga"),
								Small(
									Class("text-muted"),
									StyleAttr("display: block; font-size: 70%;"),
									g.Text("project time tracking"),
								),
							),
						),
					),
					Div(
						Class("alert alert-warning"),
						I(Class("bi-exclamation-triangle me-2")),
						g.Textf("You are currently logged in. Please logout first before using an invite link."),
					),
					Div(
						Class("text-center"),
						A(
							Href("/logout"),
							Class("btn btn-primary me-2"),
							g.Text("Logout"),
						),
					),
				),
			),
		},
	)
}

func SignupSuccess(formModel signupFormModel) g.Node {
	return Div(
		Class("alert alert-success"),
		Role("alert"),
		g.Textf("Welcome %s, you have successfully signed up! As soon as you've confirmed your email you're ready to go.", formModel.Name),
	)
}

func (a *UserWebHandlers) SignupForm(formModel signupFormModel, errorMessage string, fieldErrors map[string]string) g.Node {
	return FormEl(
		ID("signup_form"),
		ghx.Post("/signup"),

		ghx.Target("this"),
		ghx.Swap("outerHTML"),

		g.If(
			errorMessage != "",
			Div(
				Class("alert alert-danger text-center"),
				Role("alert"),
				Span(g.Text(errorMessage)),
			),
		),
		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("name"),
				Required(),
				MinLength("5"),
				MaxLength("50"),
				Type("text"),
				Name("Name"),
				Class("form-control"),
				g.Attr("placeholder", "John Doe"),
				Value(formModel.Name),
			),
			Label(
				g.Attr("for", "name"),
				g.Text("Name"),
			),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("email"),
				ghx.Post("/signup/validate"),
				Required(),
				Type("email"),
				Name("EMail"),
				g.If(
					fieldErrors["EMail"] != "",
					Class("form-control is-invalid"),
				),
				g.If(
					fieldErrors["EMail"] == "",
					Class("form-control"),
				),
				g.Attr("placeholder", "john.doe@mail.com"),
				Value(formModel.EMail),
			),
			Label(
				g.Attr("for", "email"),
				g.Text("E-Mail"),
			),
			g.If(
				fieldErrors["EMail"] != "",
				Div(
					Class("invalid-feedback"),
					g.Text(fieldErrors["EMail"]),
				),
			),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("password"),
				Required(),
				Type("password"),
				Name("Password"),
				MinLength("8"),
				MaxLength("100"),
				Value(formModel.Password),
				g.If(
					fieldErrors["Password"] != "",
					Class("form-control is-invalid"),
				),
				g.If(
					fieldErrors["Password"] == "",
					Class("form-control"),
				),
				g.Attr("placeholder", "***"),
			),
			Label(
				g.Attr("for", "password"),
				g.Text("Password"),
			),
			g.If(
				fieldErrors["Password"] != "",
				Div(
					Class("invalid-feedback"),
					g.Text(fieldErrors["Password"]),
				),
			),
		),
		Div(
			Class("form-check mb-3"),
			Input(
				ID("acceptConditions"),
				Required(),
				Type("checkbox"),
				Name("AcceptConditions"),
				Class("form-check-input"),
				g.If(formModel.AcceptConditions, Value("true")),
			),
			Label(
				g.Attr("for", "acceptConditions"),
				g.Raw(
					fmt.Sprintf("Ich bin mit den <a href=\"%v\">Datenschutzbestimmungen</a> einverstanden. I accept the <a href=\"%v\">data protection rules</a>.",
						a.config.DataProtectionURL,
						a.config.DataProtectionURL,
					),
				),
			),
		),
		Div(
			Class("container-fluid text-center"),
			Button(
				Type("submit"),
				Class("btn btn-primary w-100"),
				g.Text("Create your account"),
			),
		),
		Div(
			Class("row justify-content-around mt-2"),
			Div(
				Class("col-4 text-center"),
				g.Text("Already registered? "),
				A(
					Href("/login"),
					Class("link-secondary"),
					g.Text("Sign in here."),
				),
			),
			Div(
				Class("col-4 text-center"),
			),
		),
	)
}

func (a *UserWebHandlers) InviteSignupForm(formModel signupFormModel, errorMessage string, fieldErrors map[string]string, organization *Organization) g.Node {
	return FormEl(
		ID("baralga__main_content_modal_content"),
		Class("modal-content"),
		ghx.Post("/signup"),

		Div(
			Class("modal-header"),
			H2(
				Class("modal-title"),
				g.Textf("Join %s", organization.Title),
			),
		),
		Div(
			Class("modal-body"),
			Input(
				Type("hidden"),
				Name("CSRFToken"),
				Value(formModel.CSRFToken),
			),
			Input(
				Type("hidden"),
				Name("InviteToken"),
				Value(formModel.InviteToken),
			),
			g.If(errorMessage != "",
				Div(
					Class("alert alert-danger"),
					Role("alert"),
					g.Text(errorMessage),
				),
			),
			Div(
				Class("form-floating mb-3"),
				Input(
					ID("name"),
					Required(),
					Type("text"),
					Name("Name"),
					MinLength("5"),
					MaxLength("50"),
					Value(formModel.Name),
					g.If(
						fieldErrors["Name"] != "",
						Class("form-control is-invalid"),
					),
					g.If(
						fieldErrors["Name"] == "",
						Class("form-control"),
					),
					g.Attr("placeholder", "John Doe"),
				),
				Label(
					g.Attr("for", "name"),
					g.Text("Name"),
				),
				g.If(
					fieldErrors["Name"] != "",
					Div(
						Class("invalid-feedback"),
						g.Text(fieldErrors["Name"]),
					),
				),
			),
			Div(
				Class("form-floating mb-3"),
				Input(
					ID("email"),
					Required(),
					Type("email"),
					Name("EMail"),
					Value(formModel.EMail),
					g.If(
						fieldErrors["EMail"] != "",
						Class("form-control is-invalid"),
					),
					g.If(
						fieldErrors["EMail"] == "",
						Class("form-control"),
					),
					g.Attr("placeholder", "john@example.com"),
				),
				Label(
					g.Attr("for", "email"),
					g.Text("Email"),
				),
				g.If(
					fieldErrors["EMail"] != "",
					Div(
						Class("invalid-feedback"),
						g.Text(fieldErrors["EMail"]),
					),
				),
			),
			Div(
				Class("form-floating mb-3"),
				Input(
					ID("password"),
					Required(),
					Type("password"),
					Name("Password"),
					MinLength("8"),
					MaxLength("100"),
					Value(formModel.Password),
					g.If(
						fieldErrors["Password"] != "",
						Class("form-control is-invalid"),
					),
					g.If(
						fieldErrors["Password"] == "",
						Class("form-control"),
					),
					g.Attr("placeholder", "***"),
				),
				Label(
					g.Attr("for", "password"),
					g.Text("Password"),
				),
				g.If(
					fieldErrors["Password"] != "",
					Div(
						Class("invalid-feedback"),
						g.Text(fieldErrors["Password"]),
					),
				),
			),
			Div(
				Class("form-check mb-3"),
				Input(
					ID("acceptConditions"),
					Required(),
					Type("checkbox"),
					Name("AcceptConditions"),
					Class("form-check-input"),
					g.If(formModel.AcceptConditions, Value("true")),
				),
				Label(
					g.Attr("for", "acceptConditions"),
					g.Raw(
						fmt.Sprintf("Ich bin mit den <a href=\"%v\">Datenschutzbestimmungen</a> einverstanden. I accept the <a href=\"%v\">data protection rules</a>.",
							a.config.DataProtectionURL,
							a.config.DataProtectionURL,
						),
					),
				),
			),
		),
		Div(
			Class("modal-footer"),
			Div(
				Class("d-flex justify-content-between align-items-center w-100"),
				Div(
					Class("d-flex gap-2"),
					g.If(
						a.config.GithubClientId != "",
						A(
							Class("btn btn-secondary btn-sm"),
							Href(fmt.Sprintf("/github/login/invite/%s", formModel.InviteToken)),
							I(Class("bi-github")),
							g.Text(" GitHub"),
						),
					),
					g.If(
						a.config.GoogleClientId != "",
						A(
							Class("btn btn-secondary btn-sm"),
							Href(fmt.Sprintf("/google/login/invite/%s", formModel.InviteToken)),
							I(Class("bi-google")),
							g.Text(" Google"),
						),
					),
				),
				Button(
					Type("submit"),
					Class("btn btn-primary"),
					g.Text("Join Organization"),
				),
			),
		),
	)
}

func mapSignUpFormToUser(formModel signupFormModel, encryptedPassword string) User {
	return User{
		Name:     formModel.Name,
		Username: formModel.EMail,
		EMail:    formModel.EMail,
		Password: encryptedPassword,
		Origin:   "baralga",
	}
}

func OrganizationDialog(principal *shared.Principal, formModel organizationFormModel) g.Node {
	return FormEl(
		ID("baralga__main_content_modal_content"),
		Class("modal-content"),
		ghx.Post("/organization/update"),

		Div(
			Class("modal-header"),
			H2(
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
			Input(
				Type("hidden"),
				Name("CSRFToken"),
				Value(formModel.CSRFToken),
			),
			Div(
				Class("mb-3"),
				Label(
					Class("form-label"),
					g.Attr("for", "Name"),
					g.Text("Name"),
				),
				Input(
					ID("Name"),
					Type("text"),
					Name("Name"),
					Value(formModel.Name),
					Class("form-control"),
					g.Attr("required", "required"),
					g.Attr("maxlength", "255"),
					g.If(!principal.HasRole("ROLE_ADMIN"), g.Attr("readonly", "readonly")),
					g.If(!principal.HasRole("ROLE_ADMIN"), g.Attr("placeholder", "Contact your administrator to change this")),
				),
			),
			g.If(principal.HasRole("ROLE_ADMIN"),
				Div(
					Class("mb-3"),
					Hr(),
					Div(
						Class("d-flex justify-content-between align-items-center mb-3"),
						Div(
							H5(
								Class("mb-1"),
								I(Class("bi-people me-2")),
								g.Text("Organization Invites"),
							),
							Small(
								Class("text-muted"),
								g.Text("Generate invite links for new users"),
							),
						),
						A(
							Href("/organization/invites"),
							Class("btn btn-outline-primary btn-sm"),
							I(Class("bi-arrow-right me-1")),
							g.Text("Manage All Invites"),
						),
					),
					Div(
						ID("invite-generator"),
						Div(
							Class("d-flex gap-2 mb-2"),
							Button(
								Type("button"),
								Class("btn btn-primary btn-sm"),
								ghx.Post("/organization/invites/generate"),
								ghx.Target("#invite-link-display"),
								ghx.Swap("innerHTML"),
								I(Class("bi-plus me-1")),
								g.Text("Generate Invite Link"),
							),
						),
						Div(
							ID("invite-link-display"),
							// This will be populated when invite is generated
						),
					),
				),
			),
		),
		Div(
			Class("modal-footer"),
			g.If(principal.HasRole("ROLE_ADMIN"),
				Button(
					Type("submit"),
					Class("btn btn-primary"),
					I(Class("bi-save me-2")),
					g.Text("Update"),
				),
			),
			Button(
				Type("button"),
				Class("btn btn-secondary"),
				g.Attr("data-bs-dismiss", "modal"),
				I(Class("bi-x me-2")),
				g.Text("Close"),
			),
		),
	)
}

func OrganizationInvitesPage(config *shared.Config, principal *shared.Principal, invites []*OrganizationInvite) g.Node {
	return shared.Page(
		"Organization Invites",
		"/organization/invites",
		[]g.Node{
			Div(
				Class("container-fluid"),
				Div(
					Class("d-flex justify-content-between align-items-center mb-3"),
					H1(
						Class("h3 mb-0"),
						I(Class("bi-people me-2")),
						g.Text("Organization Invites"),
					),
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
								nodes = append(nodes, InviteCard(config, invite))
							}
							return Div(append([]g.Node{Class("row")}, nodes...)...)
						}(),
					),
				),
			),
		},
	)
}

func InviteCard(config *shared.Config, invite *OrganizationInvite) g.Node {
	status := getInviteStatus(invite)
	statusClass := getStatusClass(status)

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
						Value(getInviteURL(config, invite.Token)),
						g.Attr("readonly", "readonly"),
						ID(fmt.Sprintf("invite-url-%s", invite.ID.String())),
					),
					Button(
						Type("button"),
						Class("btn btn-outline-secondary"),
						g.Attr("onclick", fmt.Sprintf("copyToClipboard(\"invite-url-%s\")", invite.ID.String())),
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

func getInviteStatus(invite *OrganizationInvite) string {
	if invite.UsedAt != nil {
		return "Used"
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return "Expired"
	}
	return "Active"
}

func getStatusClass(status string) string {
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

func getInviteURL(config *shared.Config, token string) string {
	return fmt.Sprintf("%s/signup/invite/%s", config.Webroot, token)
}

func InviteLinkDisplay(config *shared.Config, invite *OrganizationInvite) g.Node {
	return Div(
		Class("alert alert-success"),
		Div(
			Class("d-flex justify-content-between align-items-center mb-2"),
			Strong(
				I(Class("bi-check-circle me-2")),
				g.Text("Invite Link Generated!"),
			),
			Small(
				Class("text-muted"),
				g.Textf("Expires in 24 hours"),
			),
		),
		Div(
			Class("input-group"),
			Input(
				Type("text"),
				Class("form-control font-monospace"),
				Value(getInviteURL(config, invite.Token)),
				g.Attr("readonly", "readonly"),
				ID(fmt.Sprintf("dialog-invite-url-%s", invite.ID.String())),
			),
			Button(
				Type("button"),
				Class("btn btn-outline-secondary"),
				g.Attr("onclick", fmt.Sprintf("copyToClipboard('dialog-invite-url-%s')", invite.ID.String())),
				I(Class("bi-copy")),
			),
		),
		Small(
			Class("text-muted"),
			g.Text("Share this link with new users to join your organization."),
		),
	)
}

// getPrincipalFromContext safely gets the principal from context without panicking
func (a *UserWebHandlers) getPrincipalFromContext(ctx context.Context) *shared.Principal {
	defer func() {
		// Ignore panic - means no principal found
		recover()
	}()
	return shared.MustPrincipalFromContext(ctx)
}
