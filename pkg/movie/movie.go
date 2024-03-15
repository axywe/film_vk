package movie

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/axywe/filmotheka_vk/util"
)

type Movie struct {
	ID          int       `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"releaseDate"`
	Rating      float64   `json:"rating"`
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
		util.SendJSONError(w, r, "Unsupported HTTP method", http.StatusMethodNotAllowed)
	}
}

// @Summary Create a new movie
// @Security ApiKeyAuth
// @Tags Movies
// @Accept json
// @Produce json
// @Param movie body Movie true "Movie to create"
// @Success 201 {object} Movie "Movie created"
// @Failure 400 "Bad request"
// @Failure 401 {object} util.ErrorResponse "Not authorized"
// @Failure 403 {object} util.ErrorResponse "Not authorized for this action"
// @Failure 500 "Internal server error"
// @Router /movies [post]
func (h *Handler) createMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	if m.Title == "" {
		util.SendJSONError(w, r, "Title is required", http.StatusBadRequest)
		return
	} else if len([]rune(m.Title)) > 150 {
		util.SendJSONError(w, r, "Title must be less than 255 characters", http.StatusBadRequest)
		return
	}
	if len([]rune(m.Description)) > 1000 {
		util.SendJSONError(w, r, "Description must be less than 1000 characters", http.StatusBadRequest)
		return
	}
	if m.Rating < 0 || m.Rating > 10 {
		util.SendJSONError(w, r, "Rating must be between 0 and 10", http.StatusBadRequest)
		return
	}
	sqlStatement := `INSERT INTO movies (title, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id`
	id := 0
	err := h.db.QueryRow(sqlStatement, m.Title, m.Description, m.ReleaseDate, m.Rating).Scan(&id)
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	m.ID = id
	util.SendJSONResponse(w, r, m, http.StatusCreated)
}

// @Summary Update a movie
// @Security ApiKeyAuth
// @Tags Movies
// @Accept json
// @Produce json
// @Param movie body Movie true "Movie with updated information"
// @Success 200 {object} Movie "Movie updated"
// @Failure 400 "Bad request"
// @Failure 401 {object} util.ErrorResponse "Not authorized"
// @Failure 403 {object} util.ErrorResponse "Not authorized for this action"
// @Failure 500 "Internal server error"
// @Router /movies [put]
func (h *Handler) updateMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := "UPDATE movies SET"
	params := []interface{}{}
	index := 2

	if m.Title != "" {
		sqlStatement += " title = $" + strconv.Itoa(index) + ","
		params = append(params, m.Title)
		index++
	}
	if m.Description != "" {
		sqlStatement += " description = $" + strconv.Itoa(index) + ","
		params = append(params, m.Description)
		index++
	}
	if !m.ReleaseDate.IsZero() {
		sqlStatement += " release_date = $" + strconv.Itoa(index) + ","
		params = append(params, m.ReleaseDate)
		index++
	}
	if m.Rating != 0 {
		sqlStatement += " rating = $" + strconv.Itoa(index) + ","
		params = append(params, m.Rating)
	}

	sqlStatement = strings.TrimSuffix(sqlStatement, ",")

	sqlStatement += " WHERE id = $1;"

	params = append([]interface{}{m.ID}, params...)

	_, err := h.db.Exec(sqlStatement, params...)
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	util.SendJSONResponse(w, r, m, http.StatusOK)
}

// @Summary Delete a movie
// @Security ApiKeyAuth
// @Tags Movies
// @Param id query int true "Movie ID"
// @Success 200 "Movie deleted"
// @Failure 400 "Bad request"
// @Failure 401 {object} util.ErrorResponse "Not authorized"
// @Failure 403 {object} util.ErrorResponse "Not authorized for this action"
// @Failure 404 {object} util.ErrorResponse "Movie not found"
// @Failure 500 "Internal server error"
// @Router /movies [delete]
func (h *Handler) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		util.SendJSONError(w, r, "Movie ID is required", http.StatusBadRequest)
		return
	}

	sqlStatement := `DELETE FROM actor_movie WHERE movie_id = $1;`
	_, err := h.db.Exec(sqlStatement, id)
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	sqlStatement = `DELETE FROM movies WHERE id = $1;`
	result, err := h.db.Exec(sqlStatement, id)
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		util.SendJSONError(w, r, "Movie not found", http.StatusNotFound)
		return
	}

	util.SendJSONResponse(w, r, "Movie deleted", http.StatusOK)
}

// @Summary Get list of movies
// @Security ApiKeyAuth
// @Tags Movies
// @Produce json
// @Success 200 {array} Movie "List of movies"
// @Failure 401 {object} util.ErrorResponse "Not authorized"
// @Failure 404 {object} util.ErrorResponse "No actors found"
// @Failure 500 "Internal server error"
// @Router /movies [get]
func (h *Handler) getMovies(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, title, description, release_date, rating FROM movies"

	search := r.URL.Query().Get("search")
	if search != "" {
		query += fmt.Sprintf(" WHERE title ILIKE '%%%s%%'", search)
	}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy != "title" && sortBy != "rating" && sortBy != "release_date" {
		sortBy = "rating"
	}
	sortOrder := r.URL.Query().Get("sortOrder")
	if sortOrder == "desc" || (sortBy == "rating" && sortOrder != "asc") {
		sortBy += " DESC"
	} else {
		sortBy += " ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s", sortBy)
	log.Println(query)
	rows, err := h.db.Query(query)
	if err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.ReleaseDate, &m.Rating); err != nil {
			util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		movies = append(movies, m)
	}
	if err := rows.Err(); err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(movies) == 0 {
		util.SendJSONError(w, r, "No movies found", http.StatusNotFound)
		return
	}
	util.SendJSONResponse(w, r, movies, http.StatusOK)
}
