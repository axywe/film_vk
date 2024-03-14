package movie

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axywe/filmotheka/pkg/actor"
)

type Movie struct {
	ID          int             `json:"id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ReleaseDate time.Time       `json:"releaseDate"`
	Rating      float64         `json:"rating"`
	Actors      []actor.Actor   `json:"actors"`
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
		h.createMovie(w, r)
	case http.MethodPut:
		h.updateMovie(w, r)
	case http.MethodDelete:
		h.deleteMovie(w, r)
	case http.MethodGet:
		h.getMovies(w, r)
	default:
		http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO movies (title, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id`
	id := 0
	err := h.db.QueryRow(sqlStatement, m.Title, m.Description, m.ReleaseDate, m.Rating).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.ID = id
	// Assuming there's a function to associate actors with the movie, not shown here for brevity
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) updateMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `UPDATE movies SET title = $2, description = $3, release_date = $4, rating = $5 WHERE id = $1;`
	_, err := h.db.Exec(sqlStatement, m.ID, m.Title, m.Description, m.ReleaseDate, m.Rating)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

// deleteMovie handles DELETE requests to remove a movie
func (h *Handler) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	sqlStatement := `DELETE FROM movies WHERE id = $1;`
	_, err := h.db.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Movie with ID %s was deleted successfully", id)
}

// getMovies handles GET requests to fetch all movies
func (h *Handler) getMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, title, description, release_date, rating FROM movies")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.ReleaseDate, &m.Rating); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Assuming there's a function to fetch actors for each movie, not shown here for brevity
		movies = append(movies, m)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movies)
}
