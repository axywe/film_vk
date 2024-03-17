package testutils_test

import (
	"database/sql"
	"testing"

	movie "github.com/axywe/filmotheka_vk/pkg/movie"
	_ "github.com/lib/pq"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func CreateMovie(t *testing.T, db *sql.DB, movieDetails movie.Movie) int {
	movieHandler := movie.NewHandler(db)

	body, err := json.Marshal(movieDetails)
	if err != nil {
		t.Fatalf("Unable to marshal movie: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	movieHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("Failed to create movie, status code: %v", rr.Code)
	}

	var newMovie movie.Movie
	if err := json.Unmarshal(rr.Body.Bytes(), &newMovie); err != nil {
		t.Fatalf("Unable to unmarshal response: %v", err)
	}

	return newMovie.ID
}
