package api

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

func validatePassword(password string) bool {
	return len(password) >= 8
}

func validateUsername(username string) bool {
	return len(username) >= 3
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9\-]+(\.[a-z0-9\-]+)*\.[a-z]{2,}$`)

func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}
