package main

import (
	"context"
	"net/http"

	hx "github.com/baralga/htmx"
	"github.com/baralga/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/pkg/errors"
)

type signupFormModel struct {
	CSRFToken        string
	Name             string `validate:"required,min=5,max=50"`
	EMail            string `validate:"required,email"`
	Password         string `validate:"required,min=8,max=100"`
	AcceptConditions bool
}

func (a *app) signupFormValidator(incomplete bool) func(ctx context.Context, formModel signupFormModel) (map[string]string, error) {
	validator := validator.New()
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

			_, err := a.UserRepository.FindUserByUsername(ctx, formModel.EMail)
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

func (a *app) HandleSignUpConfirm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		confirmationIDParam := chi.URLParam(r, "confirmation-id")

		userID, err := a.UserRepository.FindUserIDByConfirmationID(r.Context(), confirmationIDParam)
		if errors.Is(err, ErrUserNotFound) {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}
		if err != nil {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		err = a.ConfirmUser(r.Context(), userID)
		if err != nil {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func (a *app) HandleSignUpFormValidate() http.HandlerFunc {
	validate := a.signupFormValidator(true)
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := signupFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", nil))
			return
		}

		var formModel signupFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", nil))
			return
		}

		fieldErrors, err := validate(r.Context(), formModel)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", fieldErrors))
			return
		}

		formModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, SignupForm(formModel, "", fieldErrors))
	}
}

func (a *app) HandleSignUpPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		formModel := signupFormModel{}
		formModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, SignUpPage(r.URL.Path, formModel))
	}
}

func (a *app) HandleSignUpForm() http.HandlerFunc {
	isProduction := a.isProduction()
	validate := a.signupFormValidator(false)
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := signupFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", nil))
			return
		}

		var formModel signupFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", nil))
			return
		}

		fieldErrors, err := validate(r.Context(), formModel)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, SignupForm(formModel, "", fieldErrors))
			return
		}

		user := mapSignUpFormToUser(formModel, a.EncryptPassword(formModel.Password))
		err = a.SetUpNewUser(r.Context(), &user)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		util.RenderHTML(w, SignupSuccess(formModel))
	}
}

func SignUpPage(currentPath string, formModel signupFormModel) g.Node {
	return Page(
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
					SignupForm(formModel, "", nil),
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

func SignupForm(formModel signupFormModel, errorMessage string, fieldErrors map[string]string) g.Node {
	return FormEl(
		ID("signup_form"),
		//		Method("POST"),
		hx.Post("/signup"),

		hx.Target("this"),
		hx.Swap("outerHTML"),

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
				hx.Post("/signup/validate"),
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
				g.Text("Accept all terms and conditions."),
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
					g.Text("Login here."),
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
	}
}
