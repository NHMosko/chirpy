package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NHMosko/chirpy/internal/auth"
	"github.com/NHMosko/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
	jwtSecret string
}

func (a *apiConfig) middleMetricsInc(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	}
}


func (a *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.fileserverHits.Load())
}

func (a *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if a.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	a.fileserverHits.Store(0)
	err := a.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		log.Printf("Succesfully deleted all users")
	}
}


func (a *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type userInput struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	input := userInput{}
	decodeInput(w, r, &input)
	passwd, err := auth.HashPassword(input.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type userResponse struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	user, err := a.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email: input.Email,
		HashedPassword: passwd,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userData := userResponse{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	log.Printf("New User Created")
	respondWithJSON(w, 201, userData)
}

func (a *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type userInput struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	input := userInput{}
	decodeInput(w, r, &input)

	user, err := a.dbQueries.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	check, err := auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !check {
		log.Printf("Failed log in attempt")
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, err = a.dbQueries.RegisterRefreshToken(r.Context(), database.RegisterRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(60 * 24*time.Hour),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}


	type userResponse struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	userData := userResponse{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshToken,
	}

	log.Printf("Logged in Succesfully")
	respondWithJSON(w, http.StatusOK, userData)
}

func (a *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, "jwt")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userID, err := auth.ValidateJWT(token, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	type updateInput struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	input := updateInput{}
	decodeInput(w, r, &input)
	passwd, err := auth.HashPassword(input.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := a.dbQueries.UpdateEmailAndPassword(r.Context(), database.UpdateEmailAndPasswordParams{
		Email: input.Email,
		HashedPassword: passwd,
		ID: userID,
	})

	type userResponse struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}
	userData := userResponse{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	log.Printf("User Updated")
	respondWithJSON(w, http.StatusOK, userData)
}

func (a *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, "refresh")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := a.dbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "This refresh token has been revoked")
		return
	}

	newToken, err := auth.MakeJWT(refreshToken.UserID, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type refreshOut struct {
		Token string `json:"token"`
	}
	refreshData := refreshOut{
		Token: newToken,
	}

	log.Printf("Refreshed token Succesfully")
	respondWithJSON(w, http.StatusOK, refreshData)
}

func (a *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, "refresh")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = a.dbQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("A refresh token has been revoked.")
	w.WriteHeader(204)
}	
