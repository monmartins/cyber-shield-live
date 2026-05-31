package handlers

import (
	"html/template"

	"go.mongodb.org/mongo-driver/mongo"
)

// Handler holds shared state for all HTTP handlers.
type Handler struct {
	DB   *mongo.Database
	Tmpl *template.Template
}

// New creates a new Handler.
func New(db *mongo.Database, tmpl *template.Template) *Handler {
	return &Handler{DB: db, Tmpl: tmpl}
}

// PageData is the base template data bag.
type PageData struct {
	Title    string
	Employee interface{}
	Results  interface{}
	Query    string
	Error    string
	Message  string
	Data     interface{}
}
