package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// ─────────────────────────────────────────────────────────────────────────────
// VULN-D: Exploiting NoSQL operator injection to extract unknown fields
//
// The `/api/employee` endpoint looks up an employee by username.
// The `id` parameter is parsed as JSON if it looks like an object, allowing
// MongoDB query operators to be injected as the filter value.
//
// Additionally, the handler returns the RAW bson.M document — including ALL
// fields stored in MongoDB — not just the public-facing ones. Fields like
// `resetToken`, `secretKey`, `adminLevel`, `mfaBackupCode` are never shown
// in the normal UI but are exposed when this endpoint is called directly.
//
// Combined effect:
//   1. Operator injection lets the attacker find documents they shouldn't access
//   2. The full document response exposes fields the UI never shows
//
// Why it's vulnerable:
//   - json.Unmarshal converts {"$ne": "x"} → map[string]interface{}{"$ne": "x"}
//   - bson.D placed in MongoDB filter → operator is executed server-side
//   - bson.M decode + json.Encode returns ALL fields with no field filtering
//
// Payloads:
//   # Known username — returns full document incl. hidden fields
//   /api/employee?id=admin
//
//   # Operator: find any user whose username is not "nonexistent"
//   /api/employee?id={"$ne":"nonexistent"}
//
//   # Find admin user by prefix regex
//   /api/employee?id={"$regex":"^adm"}
//
//   # Match any document (returns first employee with ALL fields)
//   /api/employee?id={"$exists":true}
//
//   # Find user that has a specific hidden field
//   /api/employee?id={"$where":"this.hasOwnProperty('secretKey')"}
// ─────────────────────────────────────────────────────────────────────────────

// ProfileAPIHandler handles GET /api/employee?id=<value>
// Returns the FULL MongoDB document as JSON (including hidden fields).
func (h *Handler) ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Parameter 'id' is required."})
		return
	}

	// ── VULNERABLE LINES ─────────────────────────────────────────────────────
	// 1. idParam is parsed as JSON. A JSON object like {"$regex": "^adm"}
	//    becomes a map[string]interface{} containing a MongoDB operator.
	// 2. This operator is used as the filter value for the `username` field.
	// 3. bson.M decode returns ALL document fields — including hidden ones.

	var filterValue interface{} = idParam
	var parsed interface{}
	if len(idParam) > 0 && idParam[0] == '{' {
		// VULNERABLE: JSON parsed and used directly as BSON filter value
		if err := json.Unmarshal([]byte(idParam), &parsed); err == nil {
			filterValue = parsed
		}
	}

	filter := bson.D{{Key: "username", Value: filterValue}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// VULNERABLE: Decoded into bson.M → ALL fields returned, including hidden ones
	var doc bson.M
	err := h.DB.Collection("employees").FindOne(ctx, filter).Decode(&doc)
	// ─────────────────────────────────────────────────────────────────────────

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found."})
		return
	}

	// Remove MongoDB internal _id for cleaner output (still exposes all other fields)
	delete(doc, "_id")
	json.NewEncoder(w).Encode(doc)
}

// ProfilePageHandler renders GET /employee?id=X — shows only public fields in UI.
// The UI deliberately hides sensitive fields, but /api/employee exposes them all.
func (h *Handler) ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Safe lookup by username (string equality only — no operator injection here)
	filter := bson.D{{Key: "username", Value: idParam}}
	var emp publicProfile
	err := h.DB.Collection("employees").FindOne(ctx, filter).Decode(&emp)

	pd := PageData{Title: "Employee Profile – PeopleDir", Query: idParam}
	if err != nil {
		pd.Error = "Employee not found."
	} else {
		pd.Employee = emp
	}
	h.Tmpl.ExecuteTemplate(w, "profile.html", pd)
}
