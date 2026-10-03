package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]string{
		"message": message,
	})
}

var ErrInvalidAuth = errors.New("invalid authorization format")

func validateAuthHeader(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidAuth
	}
	return parts[1], nil
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
