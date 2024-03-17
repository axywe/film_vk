package movie_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	movie "github.com/axywe/filmotheka_vk/pkg/movie"
	testutils "github.com/axywe/filmotheka_vk/testutils"

	_ "github.com/lib/pq"
)

func TestCreateMovie(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := movie.NewHandler(db)

	film := movie.Movie{
		Title:       "Test Movie",
		Description: "A test movie description",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}

	b, err := json.Marshal(film)
	if err != nil {
		t.Fatalf("Failed to marshal movie: %v", err)
	}

	req, err := http.NewRequest("POST", "/movies", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var m movie.Movie
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if m.Title != film.Title {
		t.Errorf("Expected movie title to be '%s', got '%s'", film.Title, m.Title)
	}
}

func TestGetMovies(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := movie.NewHandler(db)

	req, err := http.NewRequest("GET", "/movies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var movies []movie.Movie
	if err := json.NewDecoder(rr.Body).Decode(&movies); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if len(movies) == 0 {
		t.Errorf("Expected at least one movie, got 0")
	}

	req, err = http.NewRequest("GET", "/movies?search=Movie 123", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr = httptest.NewRecorder()
	handler = http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestUpdateMovie(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := movie.NewHandler(db)

	updatedMovie := movie.Movie{
		ID:          1,
		Title:       "Updated Test Movie",
		Description: "An updated test movie description",
		ReleaseDate: time.Now(),
		Rating:      8.0,
	}

	b, err := json.Marshal(updatedMovie)
	if err != nil {
		t.Fatalf("Failed to marshal updated movie: %v", err)
	}

	req, err := http.NewRequest("PUT", "/movies", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var m movie.Movie
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if m.Title != updatedMovie.Title {
		t.Errorf("Expected movie title to be '%s', got '%s'", updatedMovie.Title, m.Title)
	}
}

func createMovie(t *testing.T, h *movie.Handler, newMovie movie.Movie) (int, error) {
	body, err := json.Marshal(newMovie)
	if err != nil {
		t.Fatalf("Unable to marshal movie for creation: %v", err)
		return 0, err
	}

	createReq, err := http.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request for movie creation: %v", err)
		return 0, err
	}
	createRr := httptest.NewRecorder()

	h.ServeHTTP(createRr, createReq)

	if createRr.Code != http.StatusCreated {
		t.Fatalf("Failed to create movie, status code: %v", createRr.Code)
	}

	var createdMovie movie.Movie
	if err := json.Unmarshal(createRr.Body.Bytes(), &createdMovie); err != nil {
		t.Fatalf("Unable to unmarshal created movie: %v", err)
		return 0, err
	}

	return createdMovie.ID, nil
}

func TestDeleteMovieEnsuringDeletion(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := movie.NewHandler(db)

	// Create a new movie to ensure there is a movie to delete
	newMovie := movie.Movie{
		Title:       "Test Movie for Deletion",
		Description: "A test movie to be deleted",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}

	createdMovieID, err := createMovie(t, h, newMovie)
	if err != nil {
		t.Fatalf("Failed to create movie for deletion test: %v", err)
	}
	// Now delete the movie
	deleteReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/movies?id=%d", createdMovieID), nil)
	if err != nil {
		t.Fatalf("Unable to create request for movie deletion: %v", err)
	}
	deleteRr := httptest.NewRecorder()

	h.ServeHTTP(deleteRr, deleteReq)

	if deleteRr.Code != http.StatusOK {
		t.Errorf("Handler returned wrong status code for delete: got %v want %v", deleteRr.Code, http.StatusOK)
	}

	responseBody := deleteRr.Body.String()
	expectedResponseBody := "\"Movie deleted\""
	if strings.TrimSpace(responseBody) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("Unexpected response body: got %v want %v", responseBody, expectedResponseBody)
	}

	var movieCount int
	err = db.QueryRow("SELECT COUNT(*) FROM movies WHERE id = $1", createdMovieID).Scan(&movieCount)
	if err != nil || movieCount != 0 {
		t.Errorf("Movie was not deleted: %v", err)
	}
}
