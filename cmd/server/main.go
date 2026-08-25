package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
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

type characterPageData struct {
	Character character.Character
	Sheet     character.Sheet
	Error     string
	Saved     bool
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

	mux.HandleFunc(
		"/characters/",
		app.requireAuth(
			app.characterHandler,
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

func (app *application) characterHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	characterID, err := characterIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {

	case http.MethodGet:
		app.characterGet(
			w,
			r,
			characterID,
		)

	case http.MethodPost:
		app.characterPost(
			w,
			r,
			characterID,
		)

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

func characterIDFromPath(
	r *http.Request,
) (int64, error) {
	idString := strings.TrimPrefix(
		r.URL.Path,
		"/characters/",
	)

	if idString == "" ||
		strings.Contains(idString, "/") {
		return 0, errors.New(
			"invalid character id",
		)
	}

	id, err := strconv.ParseInt(
		idString,
		10,
		64,
	)

	if err != nil || id < 1 {
		return 0, errors.New(
			"invalid character id",
		)
	}

	return id, nil
}

func (app *application) characterGet(
	w http.ResponseWriter,
	r *http.Request,
	characterID int64,
) {
	u, ok := currentUser(r)

	if !ok {
		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	c, err := app.characters.GetByIDForUser(
		characterID,
		u.ID,
	)

	if errors.Is(
		err,
		character.ErrNotFound,
	) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf(
			"get character error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	sheet, err := character.DecodeSheet(c.Data)

	if err != nil {
		log.Printf(
			"decode character data error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	data := characterPageData{
		Character: c,
		Sheet:     sheet,
		Saved: r.URL.Query().
			Get("saved") == "1",
	}

	renderTemplate(
		w,
		"templates/character.html",
		data,
	)
}

func (app *application) characterPost(
	w http.ResponseWriter,
	r *http.Request,
	characterID int64,
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

	u, ok := currentUser(r)

	if !ok {
		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	c, err := app.characters.GetByIDForUser(
		characterID,
		u.ID,
	)

	if errors.Is(
		err,
		character.ErrNotFound,
	) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf(
			"get character error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	name := strings.TrimSpace(
		r.PostForm.Get("name"),
	)

	class := strings.TrimSpace(
		r.PostForm.Get("class"),
	)

	levelString := strings.TrimSpace(
		r.PostForm.Get("level"),
	)

	if name == "" {
		sheet, _ := character.DecodeSheet(c.Data)

		data := characterPageData{
			Character: c,
			Sheet:     sheet,
			Error:     "Укажи имя персонажа.",
		}

		renderTemplate(
			w,
			"templates/character.html",
			data,
		)

		return
	}

	level, err := strconv.Atoi(levelString)

	if err != nil {
		sheet, _ := character.DecodeSheet(c.Data)

		data := characterPageData{
			Character: c,
			Sheet:     sheet,
			Error:     "Уровень должен быть числом.",
		}

		renderTemplate(
			w,
			"templates/character.html",
			data,
		)

		return
	}

	sheet := character.Sheet{
		Player: strings.TrimSpace(
			r.PostForm.Get("player"),
		),

		Homeland: strings.TrimSpace(
			r.PostForm.Get("homeland"),
		),

		Occupation: strings.TrimSpace(
			r.PostForm.Get("occupation"),
		),

		RaceSpecies: strings.TrimSpace(
			r.PostForm.Get("race_species"),
		),

		Goal: strings.TrimSpace(
			r.PostForm.Get("goal"),
		),

		Description: strings.TrimSpace(
			r.PostForm.Get("description"),
		),

		Background: strings.TrimSpace(
			r.PostForm.Get("background"),
		),

		BackgroundDetails: strings.TrimSpace(
			r.PostForm.Get("background_details"),
		),

		Benefits: strings.TrimSpace(
			r.PostForm.Get("benefits"),
		),

		XP: strings.TrimSpace(
			r.PostForm.Get("xp"),
		),

		Attributes: character.Attributes{
			Strength: strings.TrimSpace(
				r.PostForm.Get("strength"),
			),

			StrengthMod: strings.TrimSpace(
				r.PostForm.Get("strength_mod"),
			),

			Dexterity: strings.TrimSpace(
				r.PostForm.Get("dexterity"),
			),

			DexterityMod: strings.TrimSpace(
				r.PostForm.Get("dexterity_mod"),
			),

			Constitution: strings.TrimSpace(
				r.PostForm.Get("constitution"),
			),

			ConstitutionMod: strings.TrimSpace(
				r.PostForm.Get("constitution_mod"),
			),

			Intelligence: strings.TrimSpace(
				r.PostForm.Get("intelligence"),
			),

			IntelligenceMod: strings.TrimSpace(
				r.PostForm.Get("intelligence_mod"),
			),

			Wisdom: strings.TrimSpace(
				r.PostForm.Get("wisdom"),
			),

			WisdomMod: strings.TrimSpace(
				r.PostForm.Get("wisdom_mod"),
			),

			Charisma: strings.TrimSpace(
				r.PostForm.Get("charisma"),
			),

			CharismaMod: strings.TrimSpace(
				r.PostForm.Get("charisma_mod"),
			),
		},

		HP: character.HitPoints{
			Current: strings.TrimSpace(
				r.PostForm.Get("hp_current"),
			),

			Max: strings.TrimSpace(
				r.PostForm.Get("hp_max"),
			),
		},

		SystemStrain: character.SystemStrain{
			Current: strings.TrimSpace(
				r.PostForm.Get("system_strain_current"),
			),

			Max: strings.TrimSpace(
				r.PostForm.Get("system_strain_max"),
			),
		},

		Saves: character.Saves{
			Physical: strings.TrimSpace(
				r.PostForm.Get("save_physical"),
			),

			Evasion: strings.TrimSpace(
				r.PostForm.Get("save_evasion"),
			),

			Mental: strings.TrimSpace(
				r.PostForm.Get("save_mental"),
			),

			Luck: strings.TrimSpace(
				r.PostForm.Get("save_luck"),
			),
		},

		Combat: character.Combat{
			BaseAttackBonus: strings.TrimSpace(
				r.PostForm.Get("base_attack_bonus"),
			),

			MeleeAttack: strings.TrimSpace(
				r.PostForm.Get("melee_attack"),
			),

			RangedAttack: strings.TrimSpace(
				r.PostForm.Get("ranged_attack"),
			),

			Initiative: strings.TrimSpace(
				r.PostForm.Get("initiative"),
			),
		},

		Armor: character.Armor{
			DexMod: strings.TrimSpace(
				r.PostForm.Get("armor_dex_mod"),
			),

			WornArmor: strings.TrimSpace(
				r.PostForm.Get("worn_armor"),
			),

			AC: strings.TrimSpace(
				r.PostForm.Get("armor_class"),
			),

			Special: strings.TrimSpace(
				r.PostForm.Get("armor_special"),
			),
		},

		Skills: character.Skills{
			Administer: r.PostForm.Get("skill_administer"),
			Connect:    r.PostForm.Get("skill_connect"),
			Convince:   r.PostForm.Get("skill_convince"),
			Craft:      r.PostForm.Get("skill_craft"),
			Exert:      r.PostForm.Get("skill_exert"),
			Heal:       r.PostForm.Get("skill_heal"),
			Know:       r.PostForm.Get("skill_know"),
			Lead:       r.PostForm.Get("skill_lead"),
			Magic:      r.PostForm.Get("skill_magic"),
			Notice:     r.PostForm.Get("skill_notice"),
			Perform:    r.PostForm.Get("skill_perform"),
			Pray:       r.PostForm.Get("skill_pray"),
			Punch:      r.PostForm.Get("skill_punch"),
			Ride:       r.PostForm.Get("skill_ride"),
			Sail:       r.PostForm.Get("skill_sail"),
			Shoot:      r.PostForm.Get("skill_shoot"),
			Sneak:      r.PostForm.Get("skill_sneak"),
			Stab:       r.PostForm.Get("skill_stab"),
			Survive:    r.PostForm.Get("skill_survive"),
			Trade:      r.PostForm.Get("skill_trade"),

			WorkName: strings.TrimSpace(
				r.PostForm.Get("skill_work_name"),
			),

			Work: r.PostForm.Get("skill_work"),
		},

		SkillPoints: strings.TrimSpace(
			r.PostForm.Get("skill_points"),
		),

		ExpertPoints: strings.TrimSpace(
			r.PostForm.Get("expert_points"),
		),

		ReadiedMaxLoad: strings.TrimSpace(
			r.PostForm.Get("readied_max_load"),
		),

		StowedMaxLoad: strings.TrimSpace(
			r.PostForm.Get("stowed_max_load"),
		),

		Ammunition: strings.TrimSpace(
			r.PostForm.Get("ammunition"),
		),

		SketchOrSigil: strings.TrimSpace(
			r.PostForm.Get("sketch_or_sigil"),
		),

		Property: character.Property{
			Silver: strings.TrimSpace(
				r.PostForm.Get("property_silver"),
			),

			Gold: strings.TrimSpace(
				r.PostForm.Get("property_gold"),
			),

			StoredPossessions: strings.TrimSpace(
				r.PostForm.Get("property_stored"),
			),
		},
	}

	for i := 0; i < character.DefaultFociRows; i++ {
		sheet.Foci = append(
			sheet.Foci,
			character.Focus{
				Name: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"focus_name_%d",
							i,
						),
					),
				),

				Level: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"focus_level_%d",
							i,
						),
					),
				),

				Description: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"focus_description_%d",
							i,
						),
					),
				),
			},
		)
	}

	for i := 0; i < character.MaxSpellRows; i++ {
		name := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("spell_name_%d", i),
			),
		)

		tradition := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("spell_tradition_%d", i),
			),
		)

		level := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("spell_level_%d", i),
			),
		)

		description := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("spell_description_%d", i),
			),
		)

		// Совсем пустые слоты не сохраняем.
		if name == "" &&
			tradition == "" &&
			level == "" &&
			description == "" {
			continue
		}

		sheet.Spells = append(
			sheet.Spells,
			character.Spell{
				Name:        name,
				Tradition:   tradition,
				Level:       level,
				Description: description,
			},
		)
	}

	for i := 0; i < character.MaxMagicTraditionRows; i++ {
		name := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_tradition_name_%d", i),
			),
		)

		effortMax := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_tradition_max_%d", i),
			),
		)

		effortCurrent := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_tradition_current_%d", i),
			),
		)

		if name == "" &&
			effortMax == "" &&
			effortCurrent == "" {
			continue
		}

		sheet.MagicTraditions = append(
			sheet.MagicTraditions,
			character.MagicTradition{
				Name:          name,
				EffortMax:     effortMax,
				EffortCurrent: effortCurrent,
			},
		)
	}

	for i := 0; i < character.MaxMagicArtRows; i++ {
		name := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_art_name_%d", i),
			),
		)

		tradition := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_art_tradition_%d", i),
			),
		)

		effortSpentOn := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_art_effort_%d", i),
			),
		)

		description := strings.TrimSpace(
			r.PostForm.Get(
				fmt.Sprintf("magic_art_description_%d", i),
			),
		)

		if name == "" &&
			tradition == "" &&
			effortSpentOn == "" &&
			description == "" {
			continue
		}

		sheet.MagicArts = append(
			sheet.MagicArts,
			character.MagicArt{
				Name:          name,
				Tradition:     tradition,
				EffortSpentOn: effortSpentOn,
				Description:   description,
			},
		)
	}

	for i := 0; i < character.DefaultWeaponRows; i++ {
		sheet.Weapons = append(
			sheet.Weapons,
			character.Weapon{
				Name: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"weapon_name_%d",
							i,
						),
					),
				),

				HitBonus: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"weapon_hit_%d",
							i,
						),
					),
				),

				Damage: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"weapon_damage_%d",
							i,
						),
					),
				),

				Range: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"weapon_range_%d",
							i,
						),
					),
				),

				SpecialShock: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"weapon_special_%d",
							i,
						),
					),
				),
			},
		)
	}

	for i := 0; i < character.DefaultReadiedItemRows; i++ {
		sheet.ReadiedItems = append(
			sheet.ReadiedItems,
			character.ReadiedItem{
				Name: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"item_name_%d",
							i,
						),
					),
				),

				Encumbrance: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"item_encumbrance_%d",
							i,
						),
					),
				),

				Disabled: r.PostForm.Get(
					fmt.Sprintf(
						"item_disabled_%d",
						i,
					),
				) == "1",
			},
		)
	}

	for i := 0; i < character.DefaultStowedItemRows; i++ {
		sheet.StowedItems = append(
			sheet.StowedItems,
			character.StowedItem{
				Name: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"stowed_item_name_%d",
							i,
						),
					),
				),

				Encumbrance: strings.TrimSpace(
					r.PostForm.Get(
						fmt.Sprintf(
							"stowed_item_encumbrance_%d",
							i,
						),
					),
				),

				Disabled: r.PostForm.Get(
					fmt.Sprintf(
						"stowed_item_disabled_%d",
						i,
					),
				) == "1",
			},
		)
	}

	jsonData, err := sheet.Encode()

	if err != nil {
		log.Printf(
			"encode character data error: %v",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

	err = app.characters.Update(
		characterID,
		u.ID,
		name,
		level,
		class,
		jsonData,
	)

	if errors.Is(
		err,
		character.ErrNotFound,
	) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf(
			"update character error: %v",
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
		"/characters/"+
			strconv.FormatInt(
				characterID,
				10,
			)+
			"?saved=1",
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
