package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Note represents a knowledge-base article.
type Note struct {
	Title    string   `bson:"title"`
	Category string   `bson:"category"`
	Author   string   `bson:"author"`
	Excerpt  string   `bson:"excerpt"`
	Content  string   `bson:"content"`
	Tags     []string `bson:"tags"`
	Public   bool     `bson:"public"`
}

// PageData is the base template data bag.
type PageData struct {
	Title   string
	User    string
	Notes   []Note
	Results []Note
	Query   string
	Error   string
	Message string
	Data    interface{}
}

// HomeHandler renders the CipherNote dashboard.
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := h.GetSession(r)

	// Logged-in users see all notes; anonymous users see only public ones.
	var filter bson.D
	if user != "" {
		filter = bson.D{}
	} else {
		filter = bson.D{{Key: "public", Value: true}}
	}

	opts := options.Find().SetLimit(12).SetSort(bson.D{{Key: "_id", Value: -1}})
	cur, _ := h.DB.Collection("notes").Find(ctx, filter, opts)
	var notes []Note
	if cur != nil {
		cur.All(ctx, &notes)
	}

	h.Tmpl.ExecuteTemplate(w, "home.html", PageData{
		Title: "CipherNote – Knowledge Base",
		User:  user,
		Notes: notes,
	})
}

// LogoutHandler destroys the session.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// DocsHandler renders the internal pentest reference.
func (h *Handler) DocsHandler(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "docs.html", PageData{
		Title: "Pentest Notes – CipherNote",
		User:  h.GetSession(r),
	})
}
