package main

import (
	"slices"
	"strings"
)

func cleanWords(body string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	out := ""
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			out += "****"
		} else {
			out += word
		}
		
		if i != len(words) - 1 {
			out += " "
		}
	}

	return out
}
