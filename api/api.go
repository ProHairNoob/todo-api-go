package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"todo-api/db"
	// "todo-api/api"
)

type userSignup struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type userSignin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		var user userSignup
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				sendJSONError(w, "The request was too large", http.StatusRequestEntityTooLarge)
				return
			}
			sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
		user.Username = strings.TrimSpace(user.Username)
		if !validateUsername(user.Username) {
			sendJSONError(w, "Username must have a minimum of 3 characters", http.StatusUnprocessableEntity)
			return
		}
		user.Email = strings.ToLower(user.Email)
		user.Email = strings.TrimSpace(user.Email)
		if !validateEmail(user.Email) {
			sendJSONError(w, "Invalid Email syntax", http.StatusUnprocessableEntity)
			return
		}
		if !validatePassword(user.Password) {
			sendJSONError(w, "Password must have a minimum of 8 characters", http.StatusUnprocessableEntity)
			return
		}
		hash, err := hashPassword(user.Password)
		if err != nil {
			sendJSONError(w, "Failed to process password", http.StatusInternalServerError)
			return
		}
		userID, err := db.InsertUser(dbConn, user.Username, user.Email, hash)
		if err != nil {
			if errors.Is(err, db.ErrUserExists) {
				sendJSONError(w, "Username or email already exists", http.StatusConflict)
				return
			}
			sendJSONError(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		token, err := createToken(userID)
		if err != nil {
			sendJSONError(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"token":   token,
			"user_id": userID,
		})
	}
}

func LoginHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		var user userSignin
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				sendJSONError(w, "The request was too large", http.StatusRequestEntityTooLarge)
				return
			}
			sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
		user.Email = strings.ToLower(user.Email)
		user.Email = strings.TrimSpace(user.Email)
		if !validateEmail(user.Email) {
			sendJSONError(w, "Invalid Email syntax", http.StatusUnprocessableEntity)
			return
		}
		if !validatePassword(user.Password) {
			sendJSONError(w, "Password must have a minimum of 8 characters", http.StatusUnprocessableEntity)
			return
		}
		userID, hash, err := db.GetUserByEmail(dbConn, user.Email)
		if err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				sendJSONError(w, "invalid email or password", http.StatusUnauthorized)
				return
			}
		}
		verify := verifyPassword(user.Password, hash)
		if !verify {
			sendJSONError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		token, err := createToken(userID)
		if err != nil {
			sendJSONError(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"token":   token,
			"user_id": userID,
		})
	}
}
