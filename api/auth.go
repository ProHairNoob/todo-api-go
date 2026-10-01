package api

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	jwt.RegisteredClaims
}

func createToken(userID int64) (string, error) {
	var (
		key    []byte
		token  *jwt.Token
		signed string
	)
	// Enter your own secret key this is for debug purposes
	key = []byte("secret")
	now := time.Now()
	claims := Claims{
		jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
		},
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return signed, err
}

func verifyToken(token string) bool {
	token, err := jwt.ParseWithClaims(token)
}

func shaHashPassword(password string) string {
	sha_hash := sha256.Sum256([]byte(password))
	base_hash := base64.StdEncoding.EncodeToString(sha_hash[:])
	return string(base_hash)
}

func hashPassword(password string) (string, error) {
	password = shaHashPassword(password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), err
}

func verifyPassword(password string, hash string) bool {
	password = shaHashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
