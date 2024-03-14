package actor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Actor struct {
	ID        int          `json:"id,omitempty"`
	Name      string       `json:"name"`
	Gender    string       `json:"gender"`
	Birthdate time.Time    `json:"birthdate"`
	Movies    []MovieBrief `json:"movies,omitempty"`
}

type MovieBrief struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
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

	for _, movie := range a.Movies {
		sqlStatement = `INSERT INTO actor_movie (actor_id, movie_id) VALUES ($1, $2)`
		_, err := h.db.Exec(sqlStatement, a.ID, movie.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) updateActor(w http.ResponseWriter, r *http.Request) {
	var a Actor
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fields := []string{}
	args := []interface{}{a.ID}

	if a.Name != "" {
		fields = append(fields, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, a.Name)
	}
	if a.Gender != "" {
		fields = append(fields, fmt.Sprintf("gender = $%d", len(args)+1))
		args = append(args, a.Gender)
	}
	if !a.Birthdate.IsZero() {
		fields = append(fields, fmt.Sprintf("birthdate = $%d", len(args)+1))
		args = append(args, a.Birthdate)
	}

	if len(fields) > 0 {
		sqlStatement := fmt.Sprintf("UPDATE actors SET %s WHERE id = $1;", strings.Join(fields, ", "))
		_, err := h.db.Exec(sqlStatement, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	existingMovies, err := h.getMoviesForActor(a.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existingMovieMap := make(map[int]bool)
	for _, m := range existingMovies {
		existingMovieMap[m.ID] = true
	}
	for _, movie := range a.Movies {
		if existingMovieMap[movie.ID] {
			_, err := h.db.Exec("DELETE FROM actor_movie WHERE actor_id = $1 AND movie_id = $2", a.ID, movie.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			_, err := h.db.Exec("INSERT INTO actor_movie (actor_id, movie_id) VALUES ($1, $2)", a.ID, movie.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		delete(existingMovieMap, movie.ID)
	}

	for remainingMovieID := range existingMovieMap {
		_, err := h.db.Exec("DELETE FROM actor_movie WHERE actor_id = $1 AND movie_id = $2", a.ID, remainingMovieID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	sqlStatement := `DELETE FROM actors_movie WHERE actor_id = $1;`
	_, err := h.db.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sqlStatement = `DELETE FROM actors WHERE id = $1;`
	_, err = h.db.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Actor with ID %s was deleted successfully", id)
}

func (h *Handler) getActors(w http.ResponseWriter, _ *http.Request) {
	var actors []Actor

	rows, err := h.db.Query("SELECT id, name, gender, birthdate FROM actors")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var a Actor
		err := rows.Scan(&a.ID, &a.Name, &a.Gender, &a.Birthdate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		actors = append(actors, a)
	}

	for i, actor := range actors {
		actorMovies, err := h.getMoviesForActor(actor.ID)
		if err != nil {
			log.Println("Error fetching movies for actor:", err)
			continue
		}
		actors[i].Movies = actorMovies
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(actors)
}

func (h *Handler) getMoviesForActor(actorID int) ([]MovieBrief, error) {
	var movies []MovieBrief

	sqlStatement := `SELECT m.id, m.title FROM movies m JOIN actor_movie am ON am.movie_id = m.id WHERE am.actor_id = $1;`
	rows, err := h.db.Query(sqlStatement, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m MovieBrief
		if err := rows.Scan(&m.ID, &m.Title); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, nil
}
