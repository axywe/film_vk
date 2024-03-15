package auth_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "github.com/axywe/filmotheka_vk/internal/auth"
	testutils "github.com/axywe/filmotheka_vk/testutils"
	"golang.org/x/crypto/bcrypt"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *ErrorResponse) GenerateToken(userID int, role int) (string, error) {
	return "", errors.New(m.Message)
}

func setupTestUser(db *sql.DB, t *testing.T) {
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", "testuser", string(hashedPassword), 1)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
}

func TestBrokenAuthHandler(t *testing.T) {
	db := testutils.BrokenSetupDB(t)
	defer db.Close()

	tokenGenerator := &auth.JWTTokenGenerator{}
	h := auth.NewHandler(db, tokenGenerator)

	creds := auth.Credentials{
		Username: "testuser",
		Password: "password1234",
	}

	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	req, err := http.NewRequest("POST", "/auth", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestAuthHandler(t *testing.T) {
	db := testutils.SetupDB(t)
	setupTestUser(db, t)

	defer db.Close()
	tokenGenerator := &auth.JWTTokenGenerator{}
	h := auth.NewHandler(db, tokenGenerator)

	creds := auth.Credentials{
		Username: "testuser",
		Password: "password1234",
	}

	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	req, err := http.NewRequest("POST", "/auth", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	creds = auth.Credentials{
		Username: "testuser",
		Password: "password123",
	}

	b, err = json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	req, err = http.NewRequest("POST", "/auth", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr = httptest.NewRecorder()
	handler = http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var tokenResp auth.TokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if tokenResp.Token == "" {
		t.Errorf("Expected non-empty token")
	}
	teardownTestUser(db, t)

	newCreds := auth.Credentials{
		Username: "testuser",
		Password: "password123",
	}

	b, err = json.Marshal(newCreds)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	req, err = http.NewRequest("POST", "/auth", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr = httptest.NewRecorder()
	handler = http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	req, err = http.NewRequest("PUT", "/auth", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr = httptest.NewRecorder()
	handler = http.Handler(h)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func teardownTestUser(db *sql.DB, t *testing.T) {
	_, err := db.Exec("DELETE FROM users WHERE username = $1", "testuser")
	if err != nil {
		t.Fatalf("Failed to delete test user: %v", err)
	}
}

func TestHandler_ServeHTTP_InternalServerError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "role"}).
			AddRow(1, string(hashedPassword), 1))

	handler := auth.NewHandler(db, &ErrorResponse{Code: http.StatusInternalServerError, Message: "token generation failed"})

	requestBody := bytes.NewBufferString(`{"username":"testuser","password":"password"}`)
	req, err := http.NewRequest("POST", "/auth", requestBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expected := `{"code":500,"message":"Error while signing the token"}`
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestHandler_ServeHTTP_BadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	handler := auth.NewHandler(db, &auth.JWTTokenGenerator{})

	requestBody := bytes.NewBufferString(`{"username":"testuser", "password":`)
	req, err := http.NewRequest("POST", "/auth", requestBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `{"code":400,"message":"unexpected EOF"}`
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
