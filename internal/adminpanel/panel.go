package adminpanel

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MinunJuttu/wwn_csw/internal/adminauth"
	"github.com/MinunJuttu/wwn_csw/internal/character"
)

const (
	adminCookieName   = "wwn_admin_session"
	maxLoginFailures  = 5
	loginWindow       = 10 * time.Minute
	loginLockDuration = 10 * time.Minute
)

type loginAttempt struct {
	Failures    int
	WindowStart time.Time
	LockedUntil time.Time
}

type Panel struct {
	characters    *character.Store
	sessions      *adminauth.Store
	passwordHash  [32]byte
	secureCookies bool
	templates     map[string]*template.Template

	attemptMu sync.Mutex
	attempts  map[string]loginAttempt
}

type loginPageData struct {
	Error string
}

type characterPageData struct {
	Character character.CharacterWithOwner
	Sheet     character.Sheet
	Embed     bool
}

type charactersPageData struct {
	Characters []character.CharacterWithOwner
}

type tablePageData struct {
	Characters []character.CharacterWithOwner
}

func New(
	characters *character.Store,
	password string,
	secureCookies bool,
) (*Panel, error) {
	if characters == nil {
		return nil, fmt.Errorf("admin panel: character store is required")
	}
	if password == "" {
		return nil, fmt.Errorf("admin panel: password is required")
	}

	templateFiles := map[string]string{
		"rabbithole.html":       "templates/rabbithole.html",
		"admin.html":            "templates/admin.html",
		"admin_characters.html": "templates/admin_characters.html",
		"admin_character.html":  "templates/admin_character.html",
		"admin_table.html":      "templates/admin_table.html",
	}

	templates := make(map[string]*template.Template, len(templateFiles))

	for name, path := range templateFiles {
		t, err := template.ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parse admin template %s: %w", path, err)
		}
		templates[name] = t
	}

	return &Panel{
		characters:    characters,
		sessions:      adminauth.NewStore(),
		passwordHash:  sha256.Sum256([]byte(password)),
		secureCookies: secureCookies,
		templates:     templates,
		attempts:      make(map[string]loginAttempt),
	}, nil
}

func (p *Panel) Register(mux *http.ServeMux) {
	mux.HandleFunc("/rabbithole", p.rabbithole)
	mux.HandleFunc("/admin", p.adminHome)
	mux.HandleFunc("/admin/", p.adminSubroutes)
}

func (p *Panel) rabbithole(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/rabbithole" {
		http.NotFound(w, r)
		return
	}

	if p.isAdmin(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		p.render(w, "rabbithole.html", loginPageData{})
	case http.MethodPost:
		p.rabbitholePost(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p *Panel) rabbitholePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	ip := clientIP(r)
	if wait, locked := p.loginLocked(ip); locked {
		p.render(w, "rabbithole.html", loginPageData{
			Error: fmt.Sprintf("Слишком много попыток. Попробуй снова через %s.", friendlyDuration(wait)),
		})
		return
	}

	providedHash := sha256.Sum256([]byte(r.PostForm.Get("password")))
	if subtle.ConstantTimeCompare(providedHash[:], p.passwordHash[:]) != 1 {
		p.recordLoginFailure(ip)
		p.render(w, "rabbithole.html", loginPageData{
			Error: "Неверный пароль.",
		})
		return
	}

	p.clearLoginFailures(ip)

	token, expiresAt, err := p.sessions.Create()
	if err != nil {
		log.Printf("admin session create error: %v", err)
		http.Error(w, "Не удалось создать сессию", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   p.secureCookies,
		Expires:  expiresAt,
		MaxAge:   int(adminauth.Lifetime.Seconds()),
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (p *Panel) adminHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !p.requireAdmin(w, r) {
		return
	}

	p.render(w, "admin.html", nil)
}

func (p *Panel) adminSubroutes(w http.ResponseWriter, r *http.Request) {
	if !p.requireAdmin(w, r) {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin/")
	path = strings.Trim(path, "/")

	switch path {
	case "logout":
		p.adminLogout(w, r)
		return
	case "characters":
		p.adminCharacters(w, r)
		return
	case "table":
		p.adminTable(w, r)
		return
	}

	if strings.HasPrefix(path, "characters/") {
		p.adminCharacterRoute(w, r, strings.TrimPrefix(path, "characters/"))
		return
	}

	http.NotFound(w, r)
}

func (p *Panel) adminLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie(adminCookieName); err == nil {
		p.sessions.Delete(cookie.Value)
	}

	p.clearAdminCookie(w)
	http.Redirect(w, r, "/rabbithole", http.StatusSeeOther)
}

func (p *Panel) adminCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	characters, err := p.characters.ListAllWithOwners()
	if err != nil {
		log.Printf("admin list characters error: %v", err)
		http.Error(w, "Не удалось загрузить персонажей", http.StatusInternalServerError)
		return
	}

	p.render(w, "admin_characters.html", charactersPageData{
		Characters: characters,
	})
}

func (p *Panel) adminTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	characters, err := p.characters.ListAllWithOwners()
	if err != nil {
		log.Printf("admin table characters error: %v", err)
		http.Error(w, "Не удалось загрузить персонажей", http.StatusInternalServerError)
		return
	}

	p.render(w, "admin_table.html", tablePageData{
		Characters: characters,
	})
}

func (p *Panel) adminCharacterRoute(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	characterID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || characterID <= 0 {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		p.adminCharacterView(w, r, characterID)
		return
	}

	if len(parts) == 2 && parts[1] == "delete" {
		p.adminCharacterDelete(w, r, characterID)
		return
	}

	http.NotFound(w, r)
}

func (p *Panel) adminCharacterView(w http.ResponseWriter, r *http.Request, characterID int64) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c, err := p.characters.GetByIDWithOwner(characterID)
	if err == character.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("admin get character error: %v", err)
		http.Error(w, "Не удалось загрузить персонажа", http.StatusInternalServerError)
		return
	}

	sheet, err := character.DecodeSheet(c.Data)
	if err != nil {
		log.Printf("admin decode character %d error: %v", characterID, err)
		http.Error(w, "Не удалось прочитать лист персонажа", http.StatusInternalServerError)
		return
	}

	p.render(w, "admin_character.html", characterPageData{
		Character: c,
		Sheet:     sheet,
		Embed:     r.URL.Query().Get("embed") == "1",
	})
}

func (p *Panel) adminCharacterDelete(w http.ResponseWriter, r *http.Request, characterID int64) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := p.characters.DeleteAny(characterID); err != nil {
		if err == character.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("admin delete character %d error: %v", characterID, err)
		http.Error(w, "Не удалось удалить персонажа", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/characters", http.StatusSeeOther)
}

func (p *Panel) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if p.isAdmin(r) {
		return true
	}

	p.clearAdminCookie(w)
	http.Redirect(w, r, "/rabbithole", http.StatusSeeOther)
	return false
}

func (p *Panel) isAdmin(r *http.Request) bool {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil {
		return false
	}

	return p.sessions.Valid(cookie.Value)
}

func (p *Panel) clearAdminCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   p.secureCookies,
		MaxAge:   -1,
	})
}

func (p *Panel) render(w http.ResponseWriter, name string, data any) {
	t, ok := p.templates[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("admin template %s execute error: %v", name, err)
	}
}

func (p *Panel) loginLocked(ip string) (time.Duration, bool) {
	p.attemptMu.Lock()
	defer p.attemptMu.Unlock()

	attempt, ok := p.attempts[ip]
	if !ok {
		return 0, false
	}

	now := time.Now()
	if !attempt.LockedUntil.IsZero() && now.Before(attempt.LockedUntil) {
		return time.Until(attempt.LockedUntil), true
	}

	if now.Sub(attempt.WindowStart) > loginWindow {
		delete(p.attempts, ip)
		return 0, false
	}

	return 0, false
}

func (p *Panel) recordLoginFailure(ip string) {
	p.attemptMu.Lock()
	defer p.attemptMu.Unlock()

	now := time.Now()
	attempt := p.attempts[ip]

	if attempt.WindowStart.IsZero() || now.Sub(attempt.WindowStart) > loginWindow {
		attempt = loginAttempt{
			WindowStart: now,
		}
	}

	attempt.Failures++
	if attempt.Failures >= maxLoginFailures {
		attempt.LockedUntil = now.Add(loginLockDuration)
	}

	p.attempts[ip] = attempt
}

func (p *Panel) clearLoginFailures(ip string) {
	p.attemptMu.Lock()
	delete(p.attempts, ip)
	p.attemptMu.Unlock()
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if net.ParseIP(first) != nil {
			return first
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func friendlyDuration(d time.Duration) string {
	if d < time.Minute {
		return "меньше минуты"
	}

	minutes := int(d.Round(time.Minute) / time.Minute)
	return fmt.Sprintf("%d мин.", minutes)
}
