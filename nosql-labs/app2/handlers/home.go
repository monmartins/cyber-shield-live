package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Employee is the public-facing employee model.
type Employee struct {
	Name       string   `bson:"name"`
	Username   string   `bson:"username"`
	Department string   `bson:"department"`
	Role       string   `bson:"role"`
	Bio        string   `bson:"bio"`
	Skills     []string `bson:"skills"`
}

// HomeHandler renders the directory homepage.
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cur, _ := h.DB.Collection("employees").Find(ctx, bson.D{}, opts)

	var employees []Employee
	if cur != nil {
		cur.All(ctx, &employees)
	}

	h.Tmpl.ExecuteTemplate(w, "home.html", PageData{
		Title: "PeopleDir – Employee Directory",
		Data:  employees,
	})
}

// DocsHandler renders the pentest reference.
func (h *Handler) DocsHandler(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "docs.html", PageData{
		Title: "Pentest Notes – PeopleDir",
	})
}
