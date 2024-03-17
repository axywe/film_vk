package actor_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	actor "github.com/axywe/filmotheka_vk/pkg/actor"
	movie "github.com/axywe/filmotheka_vk/pkg/movie"
	testutils "github.com/axywe/filmotheka_vk/testutils"

	_ "github.com/lib/pq"
)

func createActor(t *testing.T, db *sql.DB, actorDetails actor.Actor) int {
	actorHandler := actor.NewHandler(db)

	body, err := json.Marshal(actorDetails)
	if err != nil {
		t.Fatalf("Unable to marshal actor: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/actors", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	actorHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("Failed to create actor, status code: %v", rr.Code)
	}

	var newActor actor.Actor
	if err := json.Unmarshal(rr.Body.Bytes(), &newActor); err != nil {
		t.Fatalf("Unable to unmarshal response: %v", err)
	}

	return newActor.ID
}

func TestCreateActor(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	newMovie1 := movie.Movie{
		Title:       "Movie 1",
		Description: "First test movie",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}
	movie1ID := testutils.CreateMovie(t, db, newMovie1)

	newMovie2 := movie.Movie{
		Title:       "Movie 2",
		Description: "Second test movie",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}
	movie2ID := testutils.CreateMovie(t, db, newMovie2)

	h := actor.NewHandler(db)

	var notMovie int
	err := db.QueryRow("SELECT MAX(id) FROM movies").Scan(&notMovie)
	if err != nil {
		t.Errorf("Error searching max id from movie: %v", err)
	}
	notMovie++
	user := actor.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
		Movies: []actor.MovieBrief{
			{ID: movie1ID, Title: "Movie 1"},
			{ID: movie2ID, Title: "Movie 2"},
			{ID: notMovie, Title: "Movie 3"},
		},
	}

	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Unable to marshal actor: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/actors", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	user = actor.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
		Movies: []actor.MovieBrief{
			{ID: movie1ID, Title: "Movie 1"},
			{ID: movie2ID, Title: "Movie 2"},
		},
	}

	body, err = json.Marshal(user)
	if err != nil {
		t.Fatalf("Unable to marshal actor: %v", err)
	}

	req, err = http.NewRequest(http.MethodPost, "/actors", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr = httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var newActor actor.Actor
	if err := json.Unmarshal(rr.Body.Bytes(), &newActor); err != nil {
		t.Fatalf("Unable to unmarshal response: %v", err)
	}

	if newActor.ID == 0 {
		t.Errorf("Expected non-zero actor ID")
	}

	var actorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM actors WHERE id = $1", newActor.ID).Scan(&actorCount)
	if err != nil || actorCount != 1 {
		t.Errorf("Actor was not created: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM actor_movie WHERE actor_id = $1", newActor.ID).Scan(&actorCount)
	if err != nil || actorCount != 2 {
		t.Errorf("The actor is not associated with the films: %v", err)
	}
}

func TestGetActors(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	req, err := http.NewRequest(http.MethodGet, "/actors", nil)
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var actors []actor.Actor
	if err := json.Unmarshal(rr.Body.Bytes(), &actors); err != nil {
		t.Fatalf("Unable to unmarshal response: %v", err)
	}

	if len(actors) == 0 {
		t.Errorf("Expected non-empty list of actors")
	}
}

func TestUnknownMethod(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	req, err := http.NewRequest(http.MethodPatch, "/actors", nil)
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestUpdateActor(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	newMovie1 := movie.Movie{
		Title:       "Movie 1",
		Description: "First test movie",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}
	movie1ID := testutils.CreateMovie(t, db, newMovie1)

	newMovie2 := movie.Movie{
		Title:       "Movie 2",
		Description: "Second test movie",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}
	movie2ID := testutils.CreateMovie(t, db, newMovie2)

	newMovie3 := movie.Movie{
		Title:       "Movie 3",
		Description: "Third test movie",
		ReleaseDate: time.Now(),
		Rating:      5.0,
	}
	movie3ID := testutils.CreateMovie(t, db, newMovie3)

	user := actor.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
		Movies: []actor.MovieBrief{
			{ID: movie1ID, Title: "Movie 1"},
			{ID: movie2ID, Title: "Movie 2"},
		},
	}

	actorID := createActor(t, db, user)

	updatedActor := actor.Actor{
		ID:        actorID,
		Name:      "Jane Doe Updated",
		Gender:    "Female",
		Birthdate: time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC),
		Movies: []actor.MovieBrief{
			{ID: movie3ID, Title: "Movie 3"},
			{ID: movie2ID, Title: "Movie 2"},
		},
	}

	body, err := json.Marshal(updatedActor)
	if err != nil {
		t.Fatalf("Unable to marshal actor: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, "/actors", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var newActor actor.Actor
	if err := json.Unmarshal(rr.Body.Bytes(), &newActor); err != nil {
		t.Fatalf("Unable to unmarshal response: %v", err)
	}

	if newActor.Name != updatedActor.Name {
		t.Errorf("Expected updated actor name to be '%v', got '%v'", updatedActor.Name, newActor.Name)
	}
}

func TestDeleteActor(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	newActor := actor.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	createdActorID := createActor(t, db, newActor)

	deleteReq, err := http.NewRequest(http.MethodDelete, "/actors", nil)
	if err != nil {
		t.Fatalf("Unable to create request for actor deletion: %v", err)
	}
	deleteRr := httptest.NewRecorder()

	h.ServeHTTP(deleteRr, deleteReq)

	if deleteRr.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for delete: got %v want %v", deleteRr.Code, http.StatusBadRequest)
	}

	deleteReq, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/actors?id=%d", createdActorID), nil)
	if err != nil {
		t.Fatalf("Unable to create request for actor deletion: %v", err)
	}
	deleteRr = httptest.NewRecorder()

	h.ServeHTTP(deleteRr, deleteReq)

	if deleteRr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code for delete: got %v want %v", deleteRr.Code, http.StatusOK)
	}

	responseBody := deleteRr.Body.String()

	expectedResponseBody := "\"Actor deleted\""

	if strings.TrimSpace(responseBody) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("Unexpected response body: got %v want %v", responseBody, expectedResponseBody)
	}

	var actorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM actors WHERE id = $1", 999).Scan(&actorCount)
	if err != nil || actorCount != 0 {
		t.Errorf("Actor was not deleted: %v", err)
	}
}
