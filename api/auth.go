package api

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	jwt.RegisteredClaims
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	key, exists := os.LookupEnv("SECRET_KEY")
	key = strings.TrimSpace(key)
	if !exists {
		log.Fatal("error no secret key given in .env")
	} else if key == "" {
		log.Fatal("error secret key cannot be empty")
	}
}

var key = []byte(os.Getenv("SECRET_KEY"))

func createToken(userID int64) (string, error) {
	var (
		token  *jwt.Token
		signed string
	)
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

var (
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidSignature = errors.New("signature is invalid")
	ErrInvalidToken     = errors.New("invalid token")
)

func verifyToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			return key, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		} else if errors.Is(err, jwt.ErrSignatureInvalid) {
			return 0, ErrInvalidSignature
		} else {
			return 0, ErrInvalidToken
		}
	}
	if !token.Valid {
		return 0, ErrInvalidToken
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return userID, nil
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
