package actor_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	actor "github.com/axywe/filmotheka_vk/pkg/actor"
	testutils "github.com/axywe/filmotheka_vk/testutils"

	_ "github.com/lib/pq"
)

func TestCreateActor(t *testing.T) {
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	user := actor.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
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

func TestUpdateActor(t *testing.T) { // TODO: Add films and users
	db := testutils.SetupDB(t)
	defer db.Close()

	h := actor.NewHandler(db)

	updatedActor := actor.Actor{
		ID:        1,
		Name:      "Jane Doe Updated",
		Gender:    "Female",
		Birthdate: time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC),
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

	body, err := json.Marshal(newActor)
	if err != nil {
		t.Fatalf("Unable to marshal actor for creation: %v", err)
	}

	createReq, err := http.NewRequest(http.MethodPost, "/actors", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to create request for actor creation: %v", err)
	}
	createRr := httptest.NewRecorder()

	h.ServeHTTP(createRr, createReq)

	if createRr.Code != http.StatusCreated {
		t.Fatalf("Failed to create actor for deletion test, status code: %v", createRr.Code)
	}

	var createdActor actor.Actor
	if err := json.Unmarshal(createRr.Body.Bytes(), &createdActor); err != nil {
		t.Fatalf("Unable to unmarshal created actor: %v", err)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/actors?id=%d", createdActor.ID), nil)
	if err != nil {
		t.Fatalf("Unable to create request for actor deletion: %v", err)
	}
	deleteRr := httptest.NewRecorder()

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
