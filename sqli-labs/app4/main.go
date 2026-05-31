package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

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
		env("DB_PASSWORD", "sqlipass"), env("DB_NAME", "app4_db"),
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
	http.HandleFunc("/search", searchHandler)       // VULN-13
	http.HandleFunc("/department", departmentHandler) // VULN-14
	http.HandleFunc("/employee", employeeHandler)   // VULN-15
	http.HandleFunc("/profile", profileHandler)     // VULN-16
	http.HandleFunc("/oob-log", oobLogHandler)
	http.HandleFunc("/docs", docsHandler)

	port := env("PORT", "8084")
	log.Printf("[StaffPortal] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, rows, _ := queryRows(
		"SELECT id, name, department, title, email, hire_date FROM employees ORDER BY id LIMIT 6",
	)
	tmpl.ExecuteTemplate(w, "home.html", PageData{Title: "StaffPortal", Data: rows})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-13: Visible error-based SQL injection
// Type   : Error-based via GET parameter `name`
// Goal   : PostgreSQL CAST errors leak data directly in the HTTP response.
// Payload: ' AND 1=CAST((SELECT password FROM credentials LIMIT 1) AS INTEGER)--
//          ' AND EXTRACTVALUE(1,CONCAT(0x7e,(SELECT version())))--  (MySQL sim)
// ─────────────────────────────────────────────────────────────────────────────
func searchHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = ""
	}

	type Result struct {
		Name  string
		Cols  []string
		Rows  [][]string
		Error string
	}

	res := Result{Name: name}

	if name != "" {
		// VULNERABLE: `name` injected directly. CAST errors surfaced to UI.
		query := "SELECT id, name, department, title, email FROM employees WHERE name ILIKE '%" + name + "%'"
		cols, rows, err := queryRows(query)
		if err != nil {
			// Error message rendered in the page — visible error-based SQLi
			res.Error = "Database error: " + err.Error()
		} else {
			res.Cols = cols
			res.Rows = rows
		}
	}

	tmpl.ExecuteTemplate(w, "search.html", PageData{Title: "Employee Search", Data: res})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-14: Blind SQL injection – time delays
// Type   : Time-based blind via GET parameter `dept`
// Goal   : Confirm injection by observing delayed HTTP response.
// Payload: Engineering'||(SELECT pg_sleep(5))--
//          Engineering'; SELECT pg_sleep(5)--
// ─────────────────────────────────────────────────────────────────────────────
func departmentHandler(w http.ResponseWriter, r *http.Request) {
	dept := r.URL.Query().Get("dept")
	if dept == "" {
		dept = "Engineering"
	}

	// VULNERABLE: `dept` injected into WHERE clause; pg_sleep causes delay.
	query := "SELECT id, name, title, email, hire_date FROM employees WHERE department = '" + dept + "'"

	start := time.Now()
	cols, rows, err := queryRows(query)
	elapsed := time.Since(start)

	type Result struct {
		Dept    string
		Cols    []string
		Rows    [][]string
		Elapsed string
		Error   string
	}

	res := Result{
		Dept:    dept,
		Cols:    cols,
		Rows:    rows,
		Elapsed: fmt.Sprintf("%.3fs", elapsed.Seconds()),
	}
	if err != nil {
		res.Error = err.Error()
	}

	tmpl.ExecuteTemplate(w, "department.html", PageData{Title: "Department – " + dept, Data: res})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-15: Blind SQL injection – time delays + information retrieval
// Type   : Time-based blind via GET parameter `id` with conditional pg_sleep
// Goal   : Extract data char-by-char using response timing as oracle.
// Payload: 1; SELECT CASE WHEN SUBSTRING(password,1,1)='S' THEN pg_sleep(5) ELSE pg_sleep(0) END FROM credentials WHERE username='admin'--
// ─────────────────────────────────────────────────────────────────────────────
func employeeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// VULNERABLE: `id` concatenated. Attacker injects conditional pg_sleep.
	query := "SELECT id, name, department, title, email, salary::text, hire_date FROM employees WHERE id = " + id

	type Employee struct {
		ID, Name, Department, Title, Email, Salary, HireDate string
	}

	start := time.Now()
	row := db.QueryRow(query)
	var emp Employee
	err := row.Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Title, &emp.Email, &emp.Salary, &emp.HireDate)
	elapsed := time.Since(start)

	type Result struct {
		Emp     *Employee
		Elapsed string
		Error   string
	}

	res := Result{Elapsed: fmt.Sprintf("%.3fs", elapsed.Seconds())}
	if err == sql.ErrNoRows {
		res.Error = "Employee not found."
	} else if err != nil {
		res.Error = "Query error (check server logs)."
	} else {
		res.Emp = &emp
	}

	tmpl.ExecuteTemplate(w, "employee.html", PageData{Title: "Employee Profile", Data: res})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-16: Blind SQL injection – out-of-band interaction (simulated)
// Type   : OOB simulation via PostgreSQL function oob_exfil() which writes to oob_log
// Goal   : Simulate DNS/HTTP exfiltration; results visible at /oob-log
// Payload: 1; SELECT oob_exfil((SELECT password FROM credentials WHERE username='admin'),'employee-id')--
//          1 OR 1=(SELECT oob_exfil(version(),'vuln16'))--
// ─────────────────────────────────────────────────────────────────────────────
func profileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "profile.html", PageData{Title: "My Profile"})
		return
	}

	empID := r.FormValue("employee_id")
	if empID == "" {
		tmpl.ExecuteTemplate(w, "profile.html", PageData{Title: "My Profile", Error: "Employee ID required."})
		return
	}

	// VULNERABLE: employee_id injected into query that calls oob_exfil() when chained.
	query := "SELECT id, name, department, title FROM employees WHERE id = " + empID

	type Result struct {
		EmpID string
		Cols  []string
		Rows  [][]string
		Error string
	}

	res := Result{EmpID: empID}
	cols, rows, err := queryRows(query)
	if err != nil {
		res.Error = "Query error (check server logs)."
	} else {
		res.Cols = cols
		res.Rows = rows
	}

	tmpl.ExecuteTemplate(w, "profile.html", PageData{Title: "My Profile", Data: res})
}

func oobLogHandler(w http.ResponseWriter, r *http.Request) {
	_, rows, err := queryRows("SELECT id, payload, source, ts FROM oob_log ORDER BY id DESC LIMIT 50")
	type Result struct {
		Rows  [][]string
		Error string
	}
	res := Result{Rows: rows}
	if err != nil {
		res.Error = err.Error()
	}
	tmpl.ExecuteTemplate(w, "ooblog.html", PageData{Title: "OOB Exfil Log", Data: res})
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "docs.html", PageData{Title: "Pentest Notes – StaffPortal"})
}
