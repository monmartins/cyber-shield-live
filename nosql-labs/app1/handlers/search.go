package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ─────────────────────────────────────────────────────────────────────────────
// VULN-A: Detecting NoSQL injection
//
// The search term `q` is interpolated directly into a MongoDB $where
// JavaScript expression — no escaping or validation.
//
// Why it's vulnerable:
//   The $where operator evaluates arbitrary JavaScript on each document.
//   By injecting JS operators, the attacker can alter query semantics:
//   - Cause a syntax error (detects injection point)
//   - Short-circuit the condition (returns all documents)
//   - Use timing functions (time-based detection)
//
// Detection payloads:
//   ?q='                       → JS syntax error → distinct HTTP error response
//   ?q=x') || (1==1) && ('     → JS always true  → returns ALL notes (incl. private)
//   ?q=x') || sleep(5000) && (' → time-based detection (5 s delay)
//   ?q=x\u0027) || (1==1) && (\u0027 → unicode escape variant
// ─────────────────────────────────────────────────────────────────────────────

// SearchHandler handles GET /search?q=<term>
func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	user := h.GetSession(r)

	// ── VULNERABLE LINE ───────────────────────────────────────────────────────
	// `q` is interpolated into JavaScript without any escaping.
	// Injecting `test') || (1==1) && ('` makes the expression always true.
	jsExpr := fmt.Sprintf(
		"this.title.toLowerCase().indexOf('%s') >= 0 || this.content.toLowerCase().indexOf('%s') >= 0",
		q, q,
	)
	// ─────────────────────────────────────────────────────────────────────────

	// Anonymous users should only see public notes, but the $where injection
	// can bypass this condition (see payloads above).
	var filter bson.D
	if user == "" {
		filter = bson.D{
			{Key: "$and", Value: bson.A{
				bson.D{{Key: "$where", Value: jsExpr}},
				bson.D{{Key: "public", Value: true}},
			}},
		}
	} else {
		filter = bson.D{{Key: "$where", Value: jsExpr}}
	}

	opts := options.Find().SetLimit(30)
	cur, err := h.DB.Collection("notes").Find(ctx, filter, opts)

	var results []Note
	errMsg := ""

	if err != nil {
		// VULN-A: A MongoDB/JS error is a clear indicator of injection.
		// The error message is intentionally surfaced so the attacker can observe it.
		errMsg = fmt.Sprintf("Search could not be completed: %v", err)
	} else {
		if err2 := cur.All(ctx, &results); err2 != nil {
			errMsg = "Error loading results."
		}
	}

	h.Tmpl.ExecuteTemplate(w, "search.html", PageData{
		Title:   "Search – CipherNote",
		User:    user,
		Results: results,
		Query:   q,
		Error:   errMsg,
	})
}
