package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"peopledir/db"
	"peopledir/handlers"
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
		env("MONGO_DB", "app2_db"),
	)

	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	h := handlers.New(database, tmpl)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", h.HomeHandler)
	mux.HandleFunc("/lookup", h.LookupPageHandler) // HTML page (VULN-C)
	mux.HandleFunc("/api/lookup", h.LookupHandler) // JSON API (VULN-C)
	mux.HandleFunc("/employee", h.ProfilePageHandler)
	mux.HandleFunc("/api/employee", h.ProfileAPIHandler) // JSON API (VULN-D)
	mux.HandleFunc("/docs", h.DocsHandler)

	port := env("PORT", "8082")
	log.Printf("[PeopleDir] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
