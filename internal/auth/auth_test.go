package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswords(t *testing.T) {
	password := "banana123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("couldn't hash password: %v", err)
		return
	}

	if _, err := CheckPasswordHash(password, hash); err != nil {
		t.Errorf("couldn't match hash: %v", err)
		return
	}
}

func TestJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "banana123"
	expiresIn := 23 * time.Second

	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("couldn't make jwt: %v", err)
		return
	}

	newUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Errorf("couldn't validate jwt: %v", err)
		return
	}

	if userID != newUserID {
		t.Errorf("id lost in translation: %v", err)
		return
	}
}

func TestGetBearerToken(t *testing.T) {
	header := make(http.Header)
	tokenString := "DSaksdhjka21sda43e_34jdsf.adas930kdd=C"
	header.Add("Authorization", "Bearer " + tokenString)
	token, err := GetBearerToken(header, "jwt") 
	if err != nil {
		t.Errorf("couldn't get bearer token: %v", err)
		return
	}
	if token != tokenString {
		t.Errorf("tokens don't match: %v != %v", token, tokenString)
		return
	}

	tokenString = ""
	header.Set("Authorization", "Bearer " + tokenString)
	token, err = GetBearerToken(header, "jwt") 
	if err == nil {
		t.Errorf("should've errored with empty token")
		return
	}


	header.Del("Authorization")
	if _, err := GetBearerToken(header, "jwt"); err == nil {
		t.Errorf("should've failed with no Authorization header")
		return
	}
}
