package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
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
		env("DB_PASSWORD", "sqlipass"), env("DB_NAME", "app5_db"),
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
	http.HandleFunc("/products", productsHandler) // VULN-17
	http.HandleFunc("/stock", stockHandler)       // VULN-18
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/oob-log", oobLogHandler)
	http.HandleFunc("/docs", docsHandler)

	port := env("PORT", "8085")
	log.Printf("[StockTrack] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, rows, _ := queryRows(
		"SELECT sku, name, quantity, location, price FROM inventory ORDER BY sku",
	)
	tmpl.ExecuteTemplate(w, "home.html", PageData{Title: "StockTrack", Data: rows})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-17: Blind SQL injection – OOB data exfiltration (simulated)
// Type   : OOB via GET parameter `sku`; oob_exfil() writes to oob_log table
// Goal   : Exfiltrate supplier api_key without visible output in response
// Payload: SKU-001' OR 1=(SELECT oob_exfil((SELECT api_key FROM suppliers LIMIT 1),'sku-oob')::integer)--
//          SKU-001'; SELECT oob_exfil(current_database(),'v17')--
// Results visible at: /oob-log
// ─────────────────────────────────────────────────────────────────────────────
func productsHandler(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")

	type Result struct {
		SKU   string
		Cols  []string
		Rows  [][]string
		Error string
		Found bool
	}

	res := Result{SKU: sku}

	if sku != "" {
		// VULNERABLE: `sku` injected directly. OOB exfiltration is silent —
		// the injected SELECT oob_exfil() runs as side effect; response unchanged.
		query := "SELECT sku, name, quantity, location, price::text FROM inventory WHERE sku = '" + sku + "'"
		cols, rows, err := queryRows(query)
		if err != nil {
			// Errors suppressed — OOB is truly blind from response perspective
			log.Printf("[VULN-17] query error: %v", err)
		} else {
			res.Cols = cols
			res.Rows = rows
			res.Found = len(rows) > 0
		}
	}

	tmpl.ExecuteTemplate(w, "products.html", PageData{Title: "Product Lookup", Data: res})
}

// ─────────────────────────────────────────────────────────────────────────────
// VULN-18: SQL injection with filter bypass via XML encoding
// Type   : POST body XML; filter blocks ASCII ' but xml.Unmarshal decodes &#x27; → '
// Goal   : Bypass single-quote filter using XML entity encoding
//
// Request: POST /stock  Content-Type: application/xml
// Body:    <stockCheck><sku>SKU-001</sku></stockCheck>
//
// Filter:  Rejects bodies containing literal ' character
// Bypass:  Replace ' with &#x27; (XML numeric entity)
//
// Payload:
//   <stockCheck><sku>SKU-001&#x27; OR &#x27;1&#x27;=&#x27;1</sku></stockCheck>
//   → After unmarshal: SKU = "SKU-001' OR '1'='1"
//   → Query: SELECT ... WHERE sku = 'SKU-001' OR '1'='1'  ← all rows returned
//
// More advanced:
//   <stockCheck><sku>&#x27; UNION SELECT api_key,name,contact,location,price::text FROM suppliers--</sku></stockCheck>
// ─────────────────────────────────────────────────────────────────────────────
func stockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, `{"error":"POST only"}`)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 8192))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"cannot read body"}`)
		return
	}

	// ── FILTER: naive check for literal single-quote in raw XML ──────────────
	// Attacker bypasses by using &#x27; (XML entity) instead of '
	if strings.Contains(string(body), "'") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"invalid characters in request"}`)
		return
	}

	// ── xml.Unmarshal decodes &#x27; → ' before we ever see it ────────────────
	type StockRequest struct {
		SKU string `xml:"sku"`
	}
	var req StockRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"xml parse error: %s"}`, err.Error())
		return
	}

	if req.SKU == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"sku required"}`)
		return
	}

	// VULNERABLE: req.SKU may contain decoded ' after XML entity expansion
	query := "SELECT sku, name, quantity FROM inventory WHERE sku = '" + req.SKU + "'"

	cols, rows, queryErr := queryRows(query)
	if queryErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"query failed: %s"}`, queryErr.Error())
		return
	}

	// Build simple JSON response
	if len(rows) == 0 {
		fmt.Fprintln(w, `{"found":false,"items":[]}`)
		return
	}

	var items []string
	for _, row := range rows {
		pairs := make([]string, len(cols))
		for i, col := range cols {
			pairs[i] = fmt.Sprintf(`"%s":"%s"`, col, strings.ReplaceAll(row[i], `"`, `\"`))
		}
		items = append(items, "{"+strings.Join(pairs, ",")+"}")
	}
	fmt.Fprintf(w, `{"found":true,"items":[%s]}`, strings.Join(items, ","))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	_, rows, _ := queryRows(
		"SELECT id, sku, quantity, status, created_at FROM orders ORDER BY id DESC",
	)
	tmpl.ExecuteTemplate(w, "orders.html", PageData{Title: "Recent Orders", Data: rows})
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
	tmpl.ExecuteTemplate(w, "docs.html", PageData{Title: "Pentest Notes – StockTrack"})
}
