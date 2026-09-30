package api

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func createToken(userID int64) (string, error) {
	var (
		key    []byte
		token  *jwt.Token
		signed string
	)
	// Enter your own secret key this is for debug purposes
	key = []byte("secret")
	token = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userID,
			"iat": time.Now().Unix(),
			"exp": time.Now().Unix() + (30 * 60),
		})
	signed, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func shaHashPassword(password string) string {
	sha_hash := sha256.Sum256([]byte(password))
	base_hash := base64.StdEncoding.EncodeToString(sha_hash[:])
	return string(base_hash)
}

func hashPassword(password string) (string, error) {
	password = shaHashPassword(password)
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed_password), err
}

func verifyPassword(password string, hash string) bool {
	password = shaHashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
