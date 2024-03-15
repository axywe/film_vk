package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/axywe/filmotheka_vk/util"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Handler struct {
	db             *sql.DB
	tokenGenerator TokenGenerator
}

func NewHandler(db *sql.DB, tokenGen TokenGenerator) *Handler {
	return &Handler{
		db:             db,
		tokenGenerator: tokenGen,
	}
}

type TokenResponse struct {
	Token string `json:"token"`
}

type TokenGenerator interface {
	GenerateToken(userID int, role int) (string, error)
}

type JWTTokenGenerator struct{}

func (j *JWTTokenGenerator) GenerateToken(userID int, role int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString([]byte("your_secret_key"))
}

// @Summary Authentication Processing
// @Description Processes POST user authentication requests and generates JWT tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "User credentials"
// @Success 200 {object} TokenResponse "Successful authentication"
// @Failure 400 {object} util.ErrorResponse "Invalid input data"
// @Failure 401 {object} util.ErrorResponse "User not found or invalid credentials"
// @Failure 405 {object} util.ErrorResponse "Only the POST method is allowed"
// @Failure 500 {object} util.ErrorResponse "Server error"
// @Router /auth [post]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.SendJSONError(w, r, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		util.SendJSONError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	var user struct {
		ID       int    `json:"id"`
		Password string `json:"password"`
		Role     int    `json:"role"`
	}
	if err := h.db.QueryRow("SELECT id, password, role FROM users WHERE username = $1", creds.Username).Scan(&user.ID, &user.Password, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			util.SendJSONError(w, r, "User not found", http.StatusUnauthorized)
		} else {
			util.SendJSONError(w, r, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		util.SendJSONError(w, r, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := h.tokenGenerator.GenerateToken(user.ID, user.Role)
	if err != nil {
		util.SendJSONError(w, r, "Error while signing the token", http.StatusInternalServerError)
		return
	}

	util.SendJSONResponse(w, r, TokenResponse{Token: tokenString}, http.StatusOK)
}
