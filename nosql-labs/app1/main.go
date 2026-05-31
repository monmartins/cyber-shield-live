package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"ciphernote/db"
	"ciphernote/handlers"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	database := db.Connect(
		env("MONGO_URI", "mongodb://localhost:27017"),
		env("MONGO_DB", "app1_db"),
	)

	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	h := handlers.New(database, tmpl)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", h.HomeHandler)
	mux.HandleFunc("/search", h.SearchHandler)
	mux.HandleFunc("/login", h.LoginHandler)
	mux.HandleFunc("/logout", h.LogoutHandler)
	mux.HandleFunc("/docs", h.DocsHandler)

	port := env("PORT", "8081")
	log.Printf("[CipherNote] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
