package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

type signupFormModel struct {
	CSRFToken        string
	Name             string `validate:"required,min=5,max=50"`
	EMail            string `validate:"required,email"`
	Password         string `validate:"required,min=8,max=100"`
	AcceptConditions bool
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
		confirmationID := uuid.New()
		err = userService.SetUpNewUser(r.Context(), &user, confirmationID)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		shared.RenderHTML(w, SignupSuccess(formModel))
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

func mapSignUpFormToUser(formModel signupFormModel, encryptedPassword string) User {
	return User{
		Name:     formModel.Name,
		Username: formModel.EMail,
		EMail:    formModel.EMail,
		Password: encryptedPassword,
		Origin:   "baralga",
	}
}
