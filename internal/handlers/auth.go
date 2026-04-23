package handlers

import (
	"embed"
	"io/fs"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"conductor/internal/db"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	db           *db.DB
	secureCookie bool
	staticFS     embed.FS
	mux          *http.ServeMux
}

// New creates a Handler and registers all routes on a new ServeMux.
func New(database *db.DB, secureCookie bool, staticFS embed.FS) *Handler {
	h := &Handler{db: database, secureCookie: secureCookie, staticFS: staticFS}
	h.mux = h.routes()
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// routes registers all application routes on a new ServeMux.
func (h *Handler) routes() *http.ServeMux {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("GET /login", h.getLogin)
	mux.HandleFunc("POST /login", h.postLogin)
	mux.HandleFunc("POST /logout", h.requireAuth(h.postLogout))

	// Home
	mux.HandleFunc("GET /{$}", h.requireAuth(h.getProjects))

	// Projects
	mux.HandleFunc("GET /projects/{id}", h.requireAuth(h.getProject))
	mux.HandleFunc("POST /projects", h.requireAuth(h.postCreateProject))
	mux.HandleFunc("POST /projects/{id}", h.requireAuth(h.postUpdateProject))
	mux.HandleFunc("POST /projects/{id}/delete", h.requireAuth(h.postDeleteProject))

	// Tasks
	mux.HandleFunc("GET /tasks", h.requireAuth(h.getAllTasks))
	mux.HandleFunc("GET /tasks/mine", h.requireAuth(h.getMyTasks))
	mux.HandleFunc("POST /tasks", h.requireAuth(h.postCreateTask))
	mux.HandleFunc("POST /tasks/{id}", h.requireAuth(h.postUpdateTask))
	mux.HandleFunc("POST /tasks/{id}/delete", h.requireAuth(h.postDeleteTask))
	mux.HandleFunc("POST /tasks/{id}/toggle", h.requireAuth(h.postToggleTask))

	// HTMX partials
	mux.HandleFunc("GET /partials/tasks", h.requireAuth(h.getTaskRows))
	mux.HandleFunc("GET /partials/tasks/new", h.requireAuth(h.getTaskForm))
	mux.HandleFunc("GET /partials/tasks/{id}", h.requireAuth(h.getTaskDetail))
	mux.HandleFunc("GET /partials/tasks/{id}/edit", h.requireAuth(h.getTaskEditForm))
	mux.HandleFunc("GET /partials/projects/new", h.requireAuth(h.getProjectForm))
	mux.HandleFunc("GET /partials/projects/{id}/edit", h.requireAuth(h.getProjectEditForm))

	// Static files + healthz
	mux.HandleFunc("GET /healthz", h.getHealthz)
	staticSub, _ := fs.Sub(h.staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	return mux
}

func (h *Handler) getLogin(w http.ResponseWriter, r *http.Request) {
	// render login.templ — stub for now, wired in Task 5
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) postLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, hash, err := h.db.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_ = h.db.DeleteExpiredSessions() // Best-effort cleanup; failure is non-fatal.
	token, err := h.db.CreateSession(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days — matches DB session sliding window
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) postLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		_ = h.db.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) getHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
