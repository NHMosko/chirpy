package main

import (
	"encoding/json"
	"log"
	"net/http"
)



func respondWithError(w http.ResponseWriter, code int, message string) {
	type errorVal struct {
		Error string `json:"error"`
	}

	retErr := errorVal{Error: message}

	errDat, err := json.Marshal(retErr)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(errDat)
}

func respondWithJSON(w http.ResponseWriter, code int, rawData any) {
	data, err := json.Marshal(rawData)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func decodeInput(w http.ResponseWriter, r *http.Request, out any) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(out)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong") //
		return
	}
}
