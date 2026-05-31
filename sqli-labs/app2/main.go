package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	tmpl *template.Template
)

type PageData struct {
	Title   string
	Message string
	Error   string
	Data    interface{}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// queryRows scans all results as strings to support UNION attacks with mixed types.
func queryRows(query string) (cols []string, rows [][]string, err error) {
	r, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()
	cols, _ = r.Columns()
	for r.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if e := r.Scan(ptrs...); e != nil {
			continue
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		rows = append(rows, row)
	}
	return
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		env("DB_HOST", "localhost"), env("DB_USER", "sqliuser"),
		env("DB_PASSWORD", "sqlipass"), env("DB_NAME", "app2_db"),
		env("DB_PORT", "5432"),
	)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("DB unreachable: %v", err)
	}

	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/catalog", catalogHandler)  // VULN-5
	http.HandleFunc("/members", membersHandler)  // VULN-6
	http.HandleFunc("/books", booksHandler)      // VULN-7
	http.HandleFunc("/authors", authorsHandler)  // VULN-8
	http.HandleFunc("/docs", docsHandler)

	port := env("PORT", "8082")
	log.Printf("[LibroBase] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, rows, _ := queryRows("SELECT id, title, author, genre, year FROM books LIMIT 5")
	tmpl.ExecuteTemplate(w, "home.html", PageData{Title: "LibroBase", Data: rows})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-5: Listing database contents – non-Oracle (PostgreSQL)
// Type   : UNION injection via GET parameter `genre`
// Goal   : Enumerate tables via information_schema.tables, then dump columns
// Payload: ' UNION SELECT table_name,table_schema,NULL FROM information_schema.tables--
// ─────────────────────────────────────────────────────────────────────────────
func catalogHandler(w http.ResponseWriter, r *http.Request) {
	genre := r.URL.Query().Get("genre")
	if genre == "" {
		genre = "Technology"
	}

	// VULNERABLE: `genre` concatenated directly. 3 columns: title, author, genre.
	query := "SELECT title, author, genre FROM books WHERE genre = '" + genre + "'"

	cols, rows, err := queryRows(query)

	d := struct {
		Genre string
		Cols  []string
		Rows  [][]string
	}{genre, cols, rows}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	tmpl.ExecuteTemplate(w, "catalog.html", PageData{
		Title: "Catalog – " + genre, Data: d, Error: errMsg,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-6: Listing database contents – Oracle simulation
// Type   : UNION injection via GET parameter `id`
// Goal   : In Oracle: SELECT table_name FROM all_tables; SELECT column_name FROM all_columns
//           PostgreSQL sim: information_schema.tables / information_schema.columns
// Payload: ' UNION SELECT label,value,NULL FROM secret_data--
//           ' UNION SELECT column_name,data_type,NULL FROM information_schema.columns WHERE table_name='users'--
// ─────────────────────────────────────────────────────────────────────────────
func membersHandler(w http.ResponseWriter, r *http.Request) {
	memberID := r.URL.Query().Get("id")

	type MemberData struct {
		ID   string
		Cols []string
		Rows [][]string
	}
	d := MemberData{ID: memberID}

	if memberID != "" {
		// VULNERABLE: 3-column query, id concatenated.
		query := "SELECT name, email, membership FROM members WHERE id = " + memberID
		var err error
		d.Cols, d.Rows, err = queryRows(query)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		tmpl.ExecuteTemplate(w, "members.html", PageData{
			Title: "Member Lookup", Data: d, Error: errMsg,
		})
		return
	}

	tmpl.ExecuteTemplate(w, "members.html", PageData{Title: "Member Lookup", Data: d})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-7: UNION attack – determining number of columns
// Type   : ORDER BY / UNION NULL probing via GET parameter `title`
// Goal   : Find column count before crafting a full UNION payload
// Payload: ' ORDER BY 1--  ' ORDER BY 2--  ' ORDER BY 3--  ' ORDER BY 4-- (error)
//           ' UNION SELECT NULL,NULL,NULL--
// ─────────────────────────────────────────────────────────────────────────────
func booksHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	type BookData struct {
		Query string
		Cols  []string
		Rows  [][]string
	}
	d := BookData{Query: title}

	if title != "" {
		// VULNERABLE: 3-column result. Attacker determines this via ORDER BY probing.
		query := "SELECT title, author, year::text FROM books WHERE title ILIKE '%" + title + "%'"
		var err error
		d.Cols, d.Rows, err = queryRows(query)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		tmpl.ExecuteTemplate(w, "books.html", PageData{
			Title: "Book Search", Data: d, Error: errMsg,
		})
		return
	}

	tmpl.ExecuteTemplate(w, "books.html", PageData{Title: "Book Search", Data: d})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-8: UNION attack – finding column containing text
// Type   : UNION with type-probing via GET parameter `name`
// Goal   : Identify which UNION column accepts text values
//           Col 1 = id (integer) → 'a' fails; Col 2 = name (text) → 'a' works
// Payload: ' UNION SELECT NULL,'probe_text',NULL--   ← col 2 is text
//           ' UNION SELECT 'probe_text',NULL,NULL--  ← fails (col 1 is int)
//           ' UNION SELECT NULL,username||':'||password,NULL FROM staff--
// ─────────────────────────────────────────────────────────────────────────────
func authorsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	type AuthorData struct {
		Query string
		Cols  []string
		Rows  [][]string
	}
	d := AuthorData{Query: name}

	if name != "" {
		// VULNERABLE: returns id(int), name(text), membership(text) → 3 cols.
		// Column 1 is INTEGER → UNION with text in col 1 will error.
		// Column 2 is TEXT    → UNION with text in col 2 succeeds.
		query := "SELECT id, name, membership FROM members WHERE name ILIKE '%" + name + "%'"
		var err error
		d.Cols, d.Rows, err = queryRows(query)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		tmpl.ExecuteTemplate(w, "authors.html", PageData{
			Title: "Member/Author Directory", Data: d, Error: errMsg,
		})
		return
	}

	tmpl.ExecuteTemplate(w, "authors.html", PageData{Title: "Member/Author Directory", Data: d})
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "docs.html", PageData{Title: "Pentest Notes – LibroBase"})
}
