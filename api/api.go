package api

import (
	"database/sql"
	"encoding/json"
	//"errors"
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

func RegisterHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var user userSignup
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			sendJSONError(w, "Invalid JSON payload", http.StatusUnprocessableEntity)
			return
		}
		if !validateEmail(user.Email) {
			sendJSONError(w, "Invalid Email syntax", http.StatusUnprocessableEntity)
			return
		}
		hash, err := hashPassword(user.Password)
		if err != nil {
			sendJSONError(w, "Failed to process password", http.StatusInternalServerError)
		}
		userID, err := db.InsertUser(dbConn, user.Username, user.Email, hash)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
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
