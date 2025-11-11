package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	check, err := argon2id.ComparePasswordAndHash(password, hash)
	return check, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer: "chirpy",
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			Subject: userID.String(),
		},
	)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (any, error) {
			if "HS256" == t.Method.Alg() {
				return []byte(tokenSecret), nil
			}
			return nil, fmt.Errorf("incorrect method")
		},
	) 
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}


func GetBearerToken(headers http.Header, tokenType string) (string, error) {
	rawHeader := headers.Get("Authorization")
	if rawHeader == "" {
		return "", fmt.Errorf("Authorization Header Not Found")
	}
	token, ok := strings.CutPrefix(rawHeader, "Bearer ")
	if !ok {
		return "", fmt.Errorf("Header format not supported (missing 'Bearer ')")
	}
	if token == "" {
		return "", fmt.Errorf("Token cannot be empty")
	}
	if tokenType == "jwt" {
		if len(token) < 100 {
			return "", fmt.Errorf("This is a refresh token")
		}
	} else if tokenType == "refresh" {
		if len(token) != 64 {
			return "", fmt.Errorf("This is not a refresh token")
		}
	}
	return token, nil
}


func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}
