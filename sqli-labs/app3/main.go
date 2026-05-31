package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

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
		env("DB_PASSWORD", "sqlipass"), env("DB_NAME", "app3_db"),
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
	http.HandleFunc("/articles", articlesHandler) // VULN-9
	http.HandleFunc("/tags", tagsHandler)          // VULN-10
	http.HandleFunc("/article", articleHandler)    // VULN-11
	http.HandleFunc("/comment", commentHandler)    // VULN-12
	http.HandleFunc("/docs", docsHandler)

	port := env("PORT", "8083")
	log.Printf("[NewsHub] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, rows, _ := queryRows(
		"SELECT id, title, category, author, created_at FROM articles WHERE published = true ORDER BY created_at DESC LIMIT 5",
	)
	tmpl.ExecuteTemplate(w, "home.html", PageData{Title: "NewsHub", Data: rows})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-9: UNION attack – retrieving data from other tables
// Type   : UNION injection via GET parameter `cat`
// Goal   : Extract usernames/passwords from the `users` table
// Payload: ' UNION SELECT title,username,password,email,'col5' FROM users--
// ─────────────────────────────────────────────────────────────────────────────
func articlesHandler(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("cat")
	if cat == "" {
		cat = "Politics"
	}

	// VULNERABLE: 5-column query. Attacker uses UNION to pull from other tables.
	query := "SELECT id::text, title, category, author, created_at::text FROM articles " +
		"WHERE category = '" + cat + "' AND published = true"

	cols, rows, err := queryRows(query)

	d := struct {
		Cat  string
		Cols []string
		Rows [][]string
	}{cat, cols, rows}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	tmpl.ExecuteTemplate(w, "articles.html", PageData{Title: "Articles – " + cat, Data: d, Error: errMsg})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-10: UNION attack – retrieving multiple values in a single column
// Type   : UNION injection via GET parameter `tag`; concatenation operator ||
// Goal   : Dump multiple columns (username:password) in one string field
// Payload: ' UNION SELECT NULL,username||'~'||password||'~'||email,NULL FROM users--
// ─────────────────────────────────────────────────────────────────────────────
func tagsHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		tag = "breaking"
	}

	// VULNERABLE: 3-column query (title, author, category). 
	// Attacker concatenates multiple values into a single column.
	query := "SELECT a.title, a.author, a.category FROM articles a " +
		"JOIN article_tags at2 ON a.id = at2.article_id " +
		"JOIN tags t ON at2.tag_id = t.id " +
		"WHERE t.name = '" + tag + "'"

	cols, rows, err := queryRows(query)

	d := struct {
		Tag  string
		Cols []string
		Rows [][]string
	}{tag, cols, rows}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	tmpl.ExecuteTemplate(w, "tags.html", PageData{Title: "Tag: " + tag, Data: d, Error: errMsg})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-11: Blind SQL injection – conditional responses
// Type   : Boolean-based blind via GET parameter `id`
// Behaviour: Returns article if condition is TRUE; 404-like if FALSE.
// Payload: 1' AND '1'='1   → article shown  (TRUE condition)
//          1' AND '1'='2   → "not found"    (FALSE condition)
//          1' AND SUBSTRING(password,1,1)='N' FROM users WHERE username='admin'--
// ─────────────────────────────────────────────────────────────────────────────
func articleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// VULNERABLE: `id` concatenated. Response differs based on query truth value.
	query := "SELECT id, title, content, category, author, created_at FROM articles " +
		"WHERE published = true AND id = '" + id + "'"

	type Article struct {
		ID, Title, Content, Category, Author, Date string
	}

	var a Article
	err := db.QueryRow(query).Scan(&a.ID, &a.Title, &a.Content, &a.Category, &a.Author, &a.Date)

	if err == sql.ErrNoRows {
		// FALSE condition branch — distinguishable from TRUE
		tmpl.ExecuteTemplate(w, "article.html", PageData{
			Title: "Article Not Found",
			Data:  nil,
			Error: "This article does not exist or has been removed.",
		})
		return
	}
	if err != nil {
		tmpl.ExecuteTemplate(w, "article.html", PageData{Title: "Error", Error: "Internal error"})
		return
	}

	// TRUE condition branch — article is displayed
	tmpl.ExecuteTemplate(w, "article.html", PageData{Title: a.Title, Data: a})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-12: Blind SQL injection – conditional errors
// Type   : Error-based blind via POST parameter `article_id`
// Behaviour: Triggers a PostgreSQL CAST error when condition is TRUE.
// Payload: 1 AND 1=CAST((SELECT username FROM users LIMIT 1) AS INTEGER)
//          1 AND 1=(SELECT CASE WHEN (1=1) THEN CAST(1/0 AS TEXT) ELSE '1' END)
// ─────────────────────────────────────────────────────────────────────────────
func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		articles := []string{"1", "2", "3", "4", "5"}
		tmpl.ExecuteTemplate(w, "comment.html", PageData{Title: "Leave a Comment", Data: articles})
		return
	}

	articleID := r.FormValue("article_id")
	author := r.FormValue("author")
	content := r.FormValue("content")

	if strings.TrimSpace(content) == "" {
		tmpl.ExecuteTemplate(w, "comment.html", PageData{
			Title: "Leave a Comment", Error: "Comment cannot be empty.",
		})
		return
	}

	// VULNERABLE: article_id is injected into a subquery inside INSERT.
	// A CAST error reveals data when the injected condition is true.
	query := fmt.Sprintf(
		"INSERT INTO comments(article_id, author, content) "+
			"SELECT %s, '%s', '%s' WHERE (SELECT COUNT(*) FROM articles WHERE id = %s) > 0",
		articleID, author, content, articleID,
	)

	_, err := db.Exec(query)
	if err != nil {
		// The error message leaks DB info when injection triggers a type error.
		tmpl.ExecuteTemplate(w, "comment.html", PageData{
			Title: "Leave a Comment", Error: "Submission error: " + err.Error(),
		})
		return
	}

	tmpl.ExecuteTemplate(w, "comment.html", PageData{
		Title: "Leave a Comment", Message: "Your comment has been submitted for review.",
	})
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "docs.html", PageData{Title: "Pentest Notes – NewsHub"})
}
