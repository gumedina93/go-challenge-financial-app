package handlers

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"go-challenge-financial-chat/internal/auth"
	"go-challenge-financial-chat/internal/chat"
	"go-challenge-financial-chat/internal/database"
)

type Handlers struct {
	auth *auth.Service
	hub  *chat.Hub
	db   *database.DB
}

func New(authService *auth.Service, hub *chat.Hub, db *database.DB) *Handlers {
	return &Handlers{
		auth: authService,
		hub:  hub,
		db:   db,
	}
}

func (h *Handlers) SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	r.HandleFunc("/", h.homeHandler).Methods("GET")
	r.HandleFunc("/login", h.loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", h.registerHandler).Methods("GET", "POST")
	r.HandleFunc("/chat", h.chatHandler).Methods("GET")
	r.HandleFunc("/ws", h.websocketHandler).Methods("GET")
	r.HandleFunc("/logout", h.logoutHandler).Methods("POST")
	return r
}

func (h *Handlers) homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

func (h *Handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := h.auth.Login(username, password)
	if err != nil {
		tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
		tmpl.Execute(w, map[string]string{"Error": err.Error()})
		return
	}

	h.auth.SetSession(w, user.Username)
	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

func (h *Handlers) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
		tmpl.Execute(w, map[string]bool{"Register": true})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	err := h.auth.Register(username, password)
	if err != nil {
		tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Register": true,
			"Error":    err.Error(),
		})
		return
	}

	h.auth.SetSession(w, username)
	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

func (h *Handlers) chatHandler(w http.ResponseWriter, r *http.Request) {
	username, err := h.auth.GetSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("web/templates/chat.html"))
	tmpl.Execute(w, map[string]string{"Username": username})
}

func (h *Handlers) websocketHandler(w http.ResponseWriter, r *http.Request) {
	username, err := h.auth.GetSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.hub.HandleWebSocket(w, r, username, user.ID)
}

func (h *Handlers) logoutHandler(w http.ResponseWriter, r *http.Request) {
	h.auth.ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
