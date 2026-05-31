package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

// Handler holds shared state for all HTTP handlers.
type Handler struct {
	DB   *mongo.Database
	Tmpl *template.Template
	mu   sync.RWMutex
	sess map[string]string // token → username
}

// New creates a new Handler.
func New(db *mongo.Database, tmpl *template.Template) *Handler {
	return &Handler{
		DB:   db,
		Tmpl: tmpl,
		sess: make(map[string]string),
	}
}

// GetSession returns the logged-in username or "" if none.
func (h *Handler) GetSession(r *http.Request) string {
	c, err := r.Cookie("_session")
	if err != nil {
		return ""
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sess[c.Value]
}

// SetSession creates a session cookie for username.
func (h *Handler) SetSession(w http.ResponseWriter, username string) {
	tok := newToken()
	h.mu.Lock()
	h.sess[tok] = username
	h.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "_session",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   7200,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearSession destroys the current session.
func (h *Handler) ClearSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("_session")
	if err == nil {
		h.mu.Lock()
		delete(h.sess, c.Value)
		h.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "_session", MaxAge: -1, Path: "/"})
}

func newToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)
}
