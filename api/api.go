package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
type addTask struct {
	Title       string `json:"title"`
	Description string `json:"desc"`
}
type taskUpdate struct {
	Title       string `json:"title"`
	Description string `json:"desc"`
	Status      string `json:"status"`
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
			fmt.Println("err: ", err)
			sendJSONError(w, "server failure", http.StatusInternalServerError)
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

func TodoHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		var task addTask
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				sendJSONError(w, "The request was too large", http.StatusRequestEntityTooLarge)
				return
			}
			sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		token, err := validateAuthHeader(authHeader)
		if err != nil {
			sendJSONError(w, "Invalid authorization format, Expected 'Bearer' <token>", http.StatusUnauthorized)
			return
		}
		userID, err := verifyToken(token, []byte("secret"))
		if err != nil {
			if errors.Is(err, ErrExpiredToken) {
				sendJSONError(w, "token expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidSignature) {
				sendJSONError(w, "signature is invalid", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidToken) {
				sendJSONError(w, "invalid token", http.StatusUnauthorized)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		taskID, err := db.InsertTask(dbConn, task.Title, task.Description, userID)
		if err != nil {
			fmt.Println("err: ", err)
			sendJSONError(w, "server failure", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":     taskID,
			"title":  task.Title,
			"desc":   task.Description,
			"status": "todo",
		})
	}
}

func DeleteTodoHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		strTaskID := r.PathValue("task_id")
		taskID, err := strconv.ParseInt(strTaskID, 10, 64)
		if err != nil {
			sendJSONError(w, "invalid id format, id must be a number", http.StatusBadRequest)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		token, err := validateAuthHeader(authHeader)
		if err != nil {
			sendJSONError(w, "Invalid authorization format, Expected 'Bearer' <token>", http.StatusUnauthorized)
			return
		}
		userID, err := verifyToken(token, []byte("secret"))
		if err != nil {
			if errors.Is(err, ErrExpiredToken) {
				sendJSONError(w, "token expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidSignature) {
				sendJSONError(w, "signature is invalid", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidToken) {
				sendJSONError(w, "invalid token", http.StatusUnauthorized)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		err = db.DeleteTask(dbConn, taskID, userID)
		if err != nil {
			if errors.Is(err, db.ErrTaskNotFound) {
				sendJSONError(w, "task not found", http.StatusNotFound)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateTodoHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		strTaskID := r.PathValue("task_id")
		var task taskUpdate
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				sendJSONError(w, "The request was too large", http.StatusRequestEntityTooLarge)
				return
			}
			sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		taskID, err := strconv.ParseInt(strTaskID, 10, 64)
		if err != nil {
			sendJSONError(w, "invalid id format, id must be a number", http.StatusBadRequest)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		token, err := validateAuthHeader(authHeader)
		if err != nil {
			sendJSONError(w, "Invalid authorization format, Expected 'Bearer' <token>", http.StatusUnauthorized)
			return
		}
		userID, err := verifyToken(token, []byte("secret"))
		if err != nil {
			if errors.Is(err, ErrExpiredToken) {
				sendJSONError(w, "token expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidSignature) {
				sendJSONError(w, "signature is invalid", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidToken) {
				sendJSONError(w, "invalid token", http.StatusUnauthorized)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		err = db.UpdateTask(dbConn, task.Description, task.Title, task.Status, taskID, userID)
		if err != nil {
			if errors.Is(err, db.ErrTaskNotFound) {
				sendJSONError(w, "task not found", http.StatusNotFound)
				return
			} else if errors.Is(err, db.ErrInvalidStatus) {
				sendJSONError(w, db.ErrInvalidStatus.Error(), http.StatusUnprocessableEntity)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":     taskID,
			"title":  task.Title,
			"desc":   task.Description,
			"status": task.Status,
		})
	}
}

// todo finish GET /todos with pagination
func GetTodoHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		query := r.URL.Query()
		page, err := strconv.ParseInt(query.Get("page"), 10, 64)
		if err != nil || page == 0 {
			page = 1
		}
		limit, err := strconv.ParseInt(query.Get("limit"), 10, 64)
		if err != nil || limit == 0 {
			limit = 10
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		token, err := validateAuthHeader(authHeader)
		if err != nil {
			sendJSONError(w, "Invalid authorization format, Expected 'Bearer' <token>", http.StatusUnauthorized)
			return
		}
		userID, err := verifyToken(token, []byte("secret"))
		if err != nil {
			if errors.Is(err, ErrExpiredToken) {
				sendJSONError(w, "token expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidSignature) {
				sendJSONError(w, "signature is invalid", http.StatusUnauthorized)
				return
			} else if errors.Is(err, ErrInvalidToken) {
				sendJSONError(w, "invalid token", http.StatusUnauthorized)
				return
			} else {
				fmt.Println("err: ", err)
				sendJSONError(w, "server failure", http.StatusInternalServerError)
				return
			}
		}
		tasks, err := db.GetTask(dbConn, userID, limit, page)
		if err != nil {
			fmt.Println("err: ", err)
			sendJSONError(w, "server failure", http.StatusInternalServerError)
			return
		}
		jsonData, err := json.Marshal(tasks)
		if err != nil {
			sendJSONError(w, "server failure", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}
