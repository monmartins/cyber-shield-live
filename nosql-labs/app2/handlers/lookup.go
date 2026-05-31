package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// ─────────────────────────────────────────────────────────────────────────────
// VULN-C: Exploiting NoSQL injection to extract data
//
// The `/lookup` endpoint searches for an employee by username.
// The username parameter is interpolated directly into a MongoDB $where
// JavaScript expression, enabling boolean-based character extraction.
//
// The response differs based on whether the injected condition is true or false:
//   - TRUE  → employee profile returned (JSON 200)
//   - FALSE → "not found" (JSON 404)
//
// This binary oracle allows an attacker to extract unknown field values
// character by character using regex matching.
//
// Extraction payloads:
//   # Extract password of user "jdoe" char by char
//   ?user=jdoe' && this.password.match(/^a.*/i) || 'x'=='y
//   ?user=jdoe' && this.password.match(/^b.*/i) || 'x'=='y
//   ...repeat until true → first char known, then extend regex
//
//   # Discover unknown field "resetToken"
//   ?user=jdoe' && this.hasOwnProperty('resetToken') || 'x'=='y
//   → true if the field exists
//
//   ?user=jdoe' && this.resetToken.match(/^reset_.*/i) || 'x'=='y
//   → true if resetToken starts with "reset_"
//
// The endpoint returns only public fields (name, department, bio) —
// hidden fields (password, resetToken, etc.) are NOT in the response.
// VULN-C reveals these values indirectly through the boolean oracle.
// ─────────────────────────────────────────────────────────────────────────────

// publicProfile is what the API intentionally returns — no sensitive fields.
type publicProfile struct {
	Name       string   `bson:"name"       json:"name"`
	Username   string   `bson:"username"   json:"username"`
	Department string   `bson:"department" json:"department"`
	Role       string   `bson:"role"       json:"role"`
	Bio        string   `bson:"bio"        json:"bio"`
	Skills     []string `bson:"skills"     json:"skills"`
}

// LookupHandler handles GET /lookup?user=<value>
// Also renders the lookup page for GET /lookup (no param).
func (h *Handler) LookupHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")

	// No param → render search page
	if username == "" {
		h.Tmpl.ExecuteTemplate(w, "lookup.html", PageData{Title: "Find Employee – PeopleDir"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	// ── VULNERABLE LINE ───────────────────────────────────────────────────────
	// `username` is interpolated into JavaScript without escaping.
	// Injecting `jdoe' && this.password.match(/^a.*/) || 'x'=='y` causes
	// the query to return the document only when the condition holds,
	// leaking one bit of information per request.
	jsExpr := fmt.Sprintf("this.username == '%s'", username)
	filter := bson.D{{Key: "$where", Value: jsExpr}}
	// ─────────────────────────────────────────────────────────────────────────

	w.Header().Set("Content-Type", "application/json")

	var emp publicProfile
	err := h.DB.Collection("employees").FindOne(ctx, filter).Decode(&emp)
	if err != nil {
		// FALSE branch → no match (or error) — attacker observes this 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Employee not found."})
		return
	}

	// TRUE branch → attacker observes 200 + profile body
	json.NewEncoder(w).Encode(emp)
}

// LookupPageHandler renders the lookup HTML page.
func (h *Handler) LookupPageHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		h.Tmpl.ExecuteTemplate(w, "lookup.html", PageData{Title: "Find Employee – PeopleDir"})
		return
	}

	// Proxy to the API and show results in UI
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	jsExpr := fmt.Sprintf("this.username == '%s'", user)
	filter := bson.D{{Key: "$where", Value: jsExpr}}

	var emp publicProfile
	err := h.DB.Collection("employees").FindOne(ctx, filter).Decode(&emp)

	pd := PageData{
		Title: "Find Employee – PeopleDir",
		Query: user,
	}
	if err != nil {
		pd.Error = fmt.Sprintf("No employee found for username \"%s\".", user)
	} else {
		pd.Employee = emp
	}
	h.Tmpl.ExecuteTemplate(w, "lookup.html", pd)
}
