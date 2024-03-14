package actor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Actor struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	Birthdate time.Time `json:"birthdate"`
}

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createActor(w, r)
	case http.MethodPut:
		h.updateActor(w, r)
	case http.MethodDelete:
		h.deleteActor(w, r)
	case http.MethodGet:
		h.getActors(w, r)
	default:
		http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createActor(w http.ResponseWriter, r *http.Request) {
	var a Actor
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO actors (name, gender, birthdate) VALUES ($1, $2, $3) RETURNING id`
	id := 0
	err := h.db.QueryRow(sqlStatement, a.Name, a.Gender, a.Birthdate).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.ID = id
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateActor(w http.ResponseWriter, r *http.Request) {
	var a Actor
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `UPDATE actors SET name = $2, gender = $3, birthdate = $4 WHERE id = $1;`
	_, err := h.db.Exec(sqlStatement, a.ID, a.Name, a.Gender, a.Birthdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) deleteActor(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Actor ID is required", http.StatusBadRequest)
		return
	}

	sqlStatement := `DELETE FROM actors WHERE id = $1;`
	_, err := h.db.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Actor with ID %s was deleted successfully", id)
}

func (h *Handler) getActors(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, gender, birthdate FROM actors")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	actors := []Actor{}
	for rows.Next() {
		var a Actor
		if err := rows.Scan(&a.ID, &a.Name, &a.Gender, &a.Birthdate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		actors = append(actors, a)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(actors)
}
