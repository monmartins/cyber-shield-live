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

// PageData is the generic template context.
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

// queryRows executes a raw query and returns columns + rows as strings.
// Using interface{} scanning so UNION attacks that change column types still render.
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
		if err2 := r.Scan(ptrs...); err2 != nil {
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
		env("DB_PASSWORD", "sqlipass"), env("DB_NAME", "app1_db"),
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
	http.HandleFunc("/products", productsHandler) // VULN-1
	http.HandleFunc("/login", loginHandler)        // VULN-2
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/search", searchHandler) // VULN-3
	http.HandleFunc("/promo", promoHandler)   // VULN-4
	http.HandleFunc("/docs", docsHandler)

	port := env("PORT", "8081")
	log.Printf("[ShopFlow] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// homeHandler shows 6 featured released products.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, rows, _ := queryRows(
		"SELECT id, name, description, price, category FROM products WHERE released = 1 LIMIT 6",
	)
	user, _ := r.Cookie("sf_user")
	userName := ""
	if user != nil {
		userName = user.Value
	}
	tmpl.ExecuteTemplate(w, "home.html", PageData{
		Title: "ShopFlow – Home",
		Data:  struct{ Rows [][]string; User string }{rows, userName},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-1: SQL Injection in WHERE clause – hidden data retrieval
// Type   : Classic string-based injection via GET parameter `category`
// Context: The query restricts results with `AND released = 1`, hiding
//          unreleased/internal products (released = 0).
// Payload: ?category=Electronics' OR 1=1--
//          ?category=Electronics' OR released=0--
// ─────────────────────────────────────────────────────────────────────────────
func productsHandler(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	if cat == "" {
		cat = "Electronics"
	}

	// VULNERABLE: `cat` is concatenated without sanitisation.
	query := "SELECT id, name, description, price, category FROM products " +
		"WHERE category = '" + cat + "' AND released = 1"

	cols, rows, err := queryRows(query)

	d := struct {
		Category string
		Cols     []string
		Rows     [][]string
	}{cat, cols, rows}

	errMsg := ""
	if err != nil {
		errMsg = "Unable to load products at this time."
	}
	tmpl.ExecuteTemplate(w, "products.html", PageData{
		Title: "Products – " + cat, Data: d, Error: errMsg,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-2: SQL Injection – login bypass
// Type   : Authentication bypass via POST parameter `username`
// Payload: username=admin'--   password=<anything>
//          username=' OR '1'='1'--
// ─────────────────────────────────────────────────────────────────────────────
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "login.html", PageData{Title: "Sign In"})
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	// VULNERABLE: both fields directly concatenated.
	query := "SELECT id, username, email, role FROM users " +
		"WHERE username = '" + username + "' AND password = '" + password + "'"

	var id int
	var uname, email, role string
	err := db.QueryRow(query).Scan(&id, &uname, &email, &role)
	if err == sql.ErrNoRows || err != nil {
		tmpl.ExecuteTemplate(w, "login.html", PageData{
			Title: "Sign In", Error: "Invalid username or password.",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "sf_user", Value: uname, Path: "/"})
	http.Redirect(w, r, "/?welcome=1", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "sf_user", Value: "", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/", http.StatusFound)
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-3: UNION-based DB version enumeration (simulated Oracle target)
// Type   : UNION injection via GET parameter `q`
// Oracle : ' UNION SELECT banner,NULL FROM v$version--
//           ' UNION SELECT banner,NULL FROM v$version WHERE ROWNUM=1--
// PG sim : ' UNION SELECT version(),NULL--
//           ' UNION SELECT current_database(),NULL--
// ─────────────────────────────────────────────────────────────────────────────
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	type SearchData struct {
		Query string
		Cols  []string
		Rows  [][]string
		Error string
	}

	d := SearchData{Query: q}
	if q != "" {
		// VULNERABLE: `q` is interpolated directly.
		query := "SELECT name, description FROM products " +
			"WHERE (name LIKE '%" + q + "%' OR description LIKE '%" + q + "%') " +
			"AND released = 1"
		var e error
		d.Cols, d.Rows, e = queryRows(query)
		if e != nil {
			d.Error = e.Error()
		}
	}

	tmpl.ExecuteTemplate(w, "search.html", PageData{Title: "Search", Data: d})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-4: UNION-based DB version enumeration (simulated MySQL/Microsoft target)
// Type   : UNION injection via POST parameter `code`
// MySQL  : ' UNION SELECT @@version,'1'--
// MSSQL  : ' UNION SELECT @@version,'1'--
// PG sim : ' UNION SELECT version(),'1'--
//           ' UNION SELECT user(),'1'--  → current_user in PG
// ─────────────────────────────────────────────────────────────────────────────
func promoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "promo.html", PageData{Title: "Promotions"})
		return
	}

	code := r.FormValue("code")

	// VULNERABLE: POST param `code` concatenated directly.
	// discount cast to text so UNION with text values still renders.
	query := "SELECT code, discount::text FROM promo_codes " +
		"WHERE code = '" + code + "' AND active = true"

	cols, rows, err := queryRows(query)

	d := struct {
		Cols []string
		Rows [][]string
	}{cols, rows}

	msg, errMsg := "", ""
	if err != nil {
		errMsg = "Error processing promo code."
	} else if len(rows) == 0 {
		errMsg = "Promo code not found or expired."
	} else {
		msg = fmt.Sprintf("Success! %s%% discount applied to your cart.", rows[0][1])
	}

	tmpl.ExecuteTemplate(w, "promo.html", PageData{
		Title: "Promotions", Data: d, Message: msg, Error: errMsg,
	})
}

// docsHandler renders the exploitation documentation.
func docsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "docs.html", PageData{Title: "Pentest Notes – ShopFlow"})
}
