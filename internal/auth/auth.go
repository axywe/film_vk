package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Credentials represents the username and password provided by a user for authentication
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// @Summary Authentication Processing
// @Description Processes POST user authentication requests and generates JWT tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "User credentials"
// @Success 200 {object} map[string]string "Successful authentication"
// @Failure 400 {string} string "Invalid input data"
// @Failure 401 {string} string "User not found or invalid credentials"
// @Failure 405 {string} string "Only the POST method is allowed"
// @Failure 500 {string} string "Server error"
// @Router /auth [post]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user struct {
		ID       int    `json:"id"`
		Password string `json:"password"`
		Role     int    `json:"role"`
	}

	if err := h.db.QueryRow("SELECT id, password, role FROM users WHERE username = $1", creds.Username).Scan(&user.ID, &user.Password, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"role":   user.Role,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your_secret_key"))
	if err != nil {
		http.Error(w, "Error while signing the token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
