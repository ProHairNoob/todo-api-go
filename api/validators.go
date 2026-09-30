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
	if len(password) >= 8 {
		return true
	} else {
		return false
	}
}

func validateUsername(username string) bool {
	if len(username) >= 3 {
		return true
	} else {
		return false
	}
}

func validateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}
