package main

import (
	"log"
	"net/http"
	"time"

	"github.com/NHMosko/chirpy/internal/auth"
	"github.com/NHMosko/chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpResponse struct {
	Id uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}


func (a *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type chirpInput struct {
		Body string `json:"body"`
	}
	input := chirpInput{}
	decodeInput(w, r, &input)

	if len(input.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long") //
		return
	}
	cleanBody := cleanWords(input.Body)

	token, err := auth.GetBearerToken(r.Header, "jwt")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := auth.ValidateJWT(token, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := a.dbQueries.CreateChirp(r.Context(),
		database.CreateChirpParams{
			Body: cleanBody,
			UserID: user,
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirpData := convertChirp(chirp)

	log.Printf("New Chirp sent out")
	respondWithJSON(w, 201, *chirpData)
}

func (a *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := a.dbQueries.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var allChirpsData []chirpResponse

	for _, chirp := range allChirps {
		chirpData := convertChirp(chirp)
		allChirpsData = append(allChirpsData, *chirpData)
	}

	respondWithJSON(w, 200, allChirpsData)
}

func (a *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")

	chirp_id, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := a.dbQueries.GetChirpByID(r.Context(), chirp_id)
	if err != nil {
		log.Printf("Chirp Not Found! ID: %v.", chirp_id)
		respondWithError(w, 404, "Couldn't find chirp on database")
		return
	}

	log.Printf("Chirp Found! ID: %v.", chirp_id)
	chirpData := convertChirp(chirp)

	respondWithJSON(w, 200, *chirpData)
}

func convertChirp(chirp database.Chirp) *chirpResponse {
	chirpData := chirpResponse{
		Id: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserId: chirp.UserID,
	}
	return &chirpData
}
