package main

import (
	"strings"

	"github.com/google/uuid"
)

// I want to be able to ignore spaces like "AS,  KD,   5H"
func parseCardCodes(cardsParam string) []string {
	if cardsParam == "" {
		return nil
	}
	tokens := strings.Split(cardsParam, ",")
	for i, token := range tokens {
		tokens[i] = strings.TrimSpace(token)
	}
	return tokens
}

var generateUUID = func() uuid.UUID {
	return uuid.New()
}
