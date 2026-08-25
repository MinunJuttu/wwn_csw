package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/MinunJuttu/wwn_csw/internal/auth"
	"github.com/MinunJuttu/wwn_csw/internal/character"
	"github.com/MinunJuttu/wwn_csw/internal/database"
	"github.com/MinunJuttu/wwn_csw/internal/session"
	"github.com/MinunJuttu/wwn_csw/internal/user"
)

const sessionCookieName = "wwn_session"

type contextKey string

const userContextKey contextKey = "user"

type application struct {
	users      *user.Store
	sessions   *session.Store
	characters *character.Store
}

type registerPageData struct {
	Username string
	Error    string
}

type loginPageData struct {
	Username   string
	Error      string
	Registered bool
}

type charactersPageData struct {
	Username   string
	Characters []character.Character
}

type newCharacterPageData struct {
	Name  string
	Error string
}

var usernamePattern = regexp.MustCompile(
	`^[a-zA-Z0-9_-]{3,32}$`,
)

func main() {
	db, err := database.Open("./data/wwn.db")
	if err != nil {
		log.Fatalf(
			"database error: %v",
			err,
		)
	}
	defer db.Close()

	log.Println(
		"Database ready: ./data/wwn.db",
	)

	app := &application{
		users:      user.NewStore(db),
		sessions:   session.NewStore(db),
		characters: character.NewStore(db),
	}

	mux := http.NewServeMux()

	staticFiles := http.FileServer(
		http.Dir("./static"),
	)

	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			staticFiles,
		),
	)

	mux.HandleFunc(
		"/",
		homeHandler,
	)

	mux.HandleFunc(
		"/register",
		app.registerHandler,
	)

	mux.HandleFunc(
		"/login",
		app.loginHandler,
	)

	mux.HandleFunc(
		"/logout",
		app.logoutHandler,
	)

	mux.HandleFunc(
		"/characters",
		app.requireAuth(
			app.charactersHandler,
		),
	)

	mux.HandleFunc(
		"/characters/new",
		app.requireAuth(
			app.newCharacterHandler,
		),
	)

	address := ":8080"

	log.Printf(
		"Server started at http://localhost%s",
		address,
	)

	err = http.ListenAndServe(
		address,
		mux,
	)

	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	renderTemplate(
		w,
		"templates/index.html",
		nil,
	)
}

func (app *application) registerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {

	case http.MethodGet:
		renderTemplate(
			w,
			"templates/register.html",
			registerPageData{},
		)

	case http.MethodPost:
		app.registerPost(w, r)

	default:
		w.Header().Set(
			"Allow",
			"GET, POST",
		)

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func (app *application) registerPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)
		return
	}

	username := strings.TrimSpace(
		r.PostForm.Get("username"),
	)

	password := r.PostForm.Get(
		"password",
	)

	data := registerPageData{
		Username: username,
	}

	if !usernamePattern.MatchString(
		username,
	) {
		data.Error = "Логин должен содержать от 3 до 32 символов: латинские буквы, цифры, _ или -."

		renderTemplate(
			w,
			"templates/register.html",
			data,
		)

		return
	}

	if len([]byte(password)) < 4 {
		data.Error = "Пароль должен содержать минимум 4 символов."

		renderTemplate(
			w,
			"templates/register.html",
			data,
		)

		return
	}

	if len([]byte(password)) > 72 {
		data.Error = "Пароль слишком длинный."

		renderTemplate(
			w,
			"templates/register.html",
			data,
		)

		return
	}

	passwordHash, err := auth.HashPassword(
		password,
	)

	if err != nil {
		log.Printf(
			"password hash error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	_, err = app.users.Create(
		username,
		passwordHash,
	)

	if errors.Is(
		err,
		user.ErrUsernameTaken,
	) {
		data.Error = "Такой логин уже занят."

		renderTemplate(
			w,
			"templates/register.html",
			data,
		)

		return
	}

	if err != nil {
		log.Printf(
			"create user error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	http.Redirect(
		w,
		r,
		"/login?registered=1",
		http.StatusSeeOther,
	)
}

func (app *application) loginHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {

	case http.MethodGet:
		data := loginPageData{
			Registered: r.URL.Query().
				Get("registered") == "1",
		}

		renderTemplate(
			w,
			"templates/login.html",
			data,
		)

	case http.MethodPost:
		app.loginPost(w, r)

	default:
		w.Header().Set(
			"Allow",
			"GET, POST",
		)

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func (app *application) loginPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)

		return
	}

	username := strings.TrimSpace(
		r.PostForm.Get("username"),
	)

	password := r.PostForm.Get(
		"password",
	)

	data := loginPageData{
		Username: username,
	}

	u, err := app.users.GetByUsername(
		username,
	)

	if errors.Is(
		err,
		user.ErrNotFound,
	) {
		data.Error = "Неверный логин или пароль."

		renderTemplate(
			w,
			"templates/login.html",
			data,
		)

		return
	}

	if err != nil {
		log.Printf(
			"get user error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	if !auth.CheckPassword(
		u.PasswordHash,
		password,
	) {
		data.Error = "Неверный логин или пароль."

		renderTemplate(
			w,
			"templates/login.html",
			data,
		)

		return
	}

	token, expiresAt, err :=
		app.sessions.Create(u.ID)

	if err != nil {
		log.Printf(
			"create session error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false,
			Expires:  expiresAt,
			MaxAge: int(
				session.Lifetime.Seconds(),
			),
		},
	)

	http.Redirect(
		w,
		r,
		"/characters",
		http.StatusSeeOther,
	)
}

func (app *application) logoutHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set(
			"Allow",
			"POST",
		)

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	cookie, err := r.Cookie(
		sessionCookieName,
	)

	if err == nil {
		if err := app.sessions.Delete(
			cookie.Value,
		); err != nil {
			log.Printf(
				"delete session error: %v",
				err,
			)
		}
	}

	clearSessionCookie(w)

	http.Redirect(
		w,
		r,
		"/login",
		http.StatusSeeOther,
	)
}

func (app *application) charactersHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set(
			"Allow",
			"GET",
		)

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	u, ok := currentUser(r)

	if !ok {
		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	characters, err :=
		app.characters.ListByUserID(
			u.ID,
		)

	if err != nil {
		log.Printf(
			"list characters error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	data := charactersPageData{
		Username:   u.Username,
		Characters: characters,
	}

	renderTemplate(
		w,
		"templates/characters.html",
		data,
	)
}

func (app *application) newCharacterHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {

	case http.MethodGet:
		renderTemplate(
			w,
			"templates/new_character.html",
			newCharacterPageData{},
		)

	case http.MethodPost:
		app.newCharacterPost(w, r)

	default:
		w.Header().Set(
			"Allow",
			"GET, POST",
		)

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func (app *application) newCharacterPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)

		return
	}

	name := strings.TrimSpace(
		r.PostForm.Get("name"),
	)

	data := newCharacterPageData{
		Name: name,
	}

	if len([]rune(name)) < 1 {
		data.Error = "Укажи имя персонажа."

		renderTemplate(
			w,
			"templates/new_character.html",
			data,
		)

		return
	}

	if len([]rune(name)) > 100 {
		data.Error = "Имя персонажа слишком длинное."

		renderTemplate(
			w,
			"templates/new_character.html",
			data,
		)

		return
	}

	u, ok := currentUser(r)

	if !ok {
		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	_, err = app.characters.Create(
		u.ID,
		name,
	)

	if err != nil {
		log.Printf(
			"create character error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	http.Redirect(
		w,
		r,
		"/characters",
		http.StatusSeeOther,
	)
}

func (app *application) requireAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		cookie, err := r.Cookie(
			sessionCookieName,
		)

		if err != nil {
			http.Redirect(
				w,
				r,
				"/login",
				http.StatusSeeOther,
			)

			return
		}

		userID, err :=
			app.sessions.GetUserID(
				cookie.Value,
			)

		if err != nil {
			clearSessionCookie(w)

			http.Redirect(
				w,
				r,
				"/login",
				http.StatusSeeOther,
			)

			return
		}

		u, err := app.users.GetByID(
			userID,
		)

		if err != nil {
			clearSessionCookie(w)

			http.Redirect(
				w,
				r,
				"/login",
				http.StatusSeeOther,
			)

			return
		}

		ctx := context.WithValue(
			r.Context(),
			userContextKey,
			u,
		)

		next(
			w,
			r.WithContext(ctx),
		)
	}
}

func currentUser(
	r *http.Request,
) (user.User, bool) {
	u, ok := r.Context().
		Value(userContextKey).(user.User)

	return u, ok
}

func clearSessionCookie(
	w http.ResponseWriter,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     sessionCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false,
			MaxAge:   -1,
		},
	)
}

func renderTemplate(
	w http.ResponseWriter,
	path string,
	data any,
) {
	tmpl, err := template.ParseFiles(
		path,
	)

	if err != nil {
		log.Printf(
			"template parse error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	err = tmpl.Execute(
		w,
		data,
	)

	if err != nil {
		log.Printf(
			"template execute error: %v",
			err,
		)
	}
}
