package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type DeckResponse struct {
	DeckID    string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
}

type OpenDeckResponse struct {
	DeckID    string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
	Cards     []Card `json:"cards"`
}

type DrawResponse struct {
	Cards []Card `json:"cards"`
}

func handleDeck(storage DeckStorage, w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/decks/")
	logrus.WithField("path", path).Debug("New request")
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 1 && r.Method == "POST" {
		logrus.Debug("Creating deck")
		handleCreateDeck(storage, w, r)
		return
	}
	if len(pathParts) == 1 && r.Method == "GET" {
		logrus.WithFields(logrus.Fields{
			"path":      path,
			"pathParts": pathParts,
		}).Debug("Opening deck")
		handleOpenDeck(storage, w, r, pathParts[0])
		return
	}
	if len(pathParts) == 1 && r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(pathParts) == 2 && pathParts[1] == "draw" && r.Method == "POST" {
		logrus.WithFields(logrus.Fields{
			"path":      path,
			"pathParts": pathParts,
		}).Debug("Drawing from deck")
		handleDrawCards(storage, w, r, pathParts[0])
		return
	}
	if len(pathParts) == 2 && pathParts[1] == "draw" && r.Method == "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	http.NotFound(w, r)
}

func handleCreateDeck(storage DeckStorage, w http.ResponseWriter, r *http.Request) {

	shuffle := r.URL.Query().Get("shuffle") == "true"
	cardsParam := r.URL.Query().Get("cards")
	cardCodes := parseCardCodes(cardsParam)
	logrus.Debugf("Request to create a new deck shuffle=%v cards=%v", shuffle, cardsParam)

	id := generateUUID()
	deck := NewDeck(id, shuffle, cardCodes)
	storage.SaveDeck(*deck)
	logrus.WithField("deck_id", deck.ID).Debugf("Saving new deck")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := DeckResponse{
		DeckID:    deck.ID.String(),
		Shuffled:  deck.Shuffled,
		Remaining: len(deck.Cards),
	}
	json.NewEncoder(w).Encode(response)
}

func handleOpenDeck(storage DeckStorage, w http.ResponseWriter, r *http.Request, deckIDParam string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if deckIDParam == "" {
		http.Error(w, "Missing deck ID", http.StatusBadRequest)
		return
	}

	deckID, err := uuid.Parse(deckIDParam)
	if err != nil {
		http.Error(w, "Invalid deck ID", http.StatusBadRequest)
		return
	}

	deck, err := storage.GetDeck(deckID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Deck not found: %s", err), http.StatusNotFound)
		return
	}

	logrus.WithField("deck_id", deck.ID).Debug("Opening deck")

	response := OpenDeckResponse{
		DeckID:    deckID.String(),
		Shuffled:  deck.Shuffled,
		Remaining: len(deck.Cards),
		Cards:     deck.Cards,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDrawCards(storage DeckStorage, w http.ResponseWriter, r *http.Request, deckIDParam string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if deckIDParam == "" {
		http.Error(w, "Missing deck ID", http.StatusBadRequest)
		return
	}

	deckID, err := uuid.Parse(deckIDParam)
	if err != nil {
		http.Error(w, "Invalid deck ID", http.StatusBadRequest)
		return
	}

	deck, err := storage.GetDeck(deckID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Deck not found: %v", err), http.StatusNotFound)
		return
	}

	numCardsParam := r.URL.Query().Get("count")
	numCards, err := strconv.Atoi(numCardsParam)
	if err != nil || numCards < 1 {
		http.Error(w, "Invalid number of cards", http.StatusBadRequest)
		return
	}
	if numCards > len(deck.Cards) {
		http.Error(w, "Not enough cards in the deck", http.StatusBadRequest)
		return
	}

	logrus.WithField("deck_id", deck.ID).Debugf("Drawing count=%v cards from deck", numCards)
	drawnCards := deck.Draw(numCards)
	storage.UpdateDeck(deck)
	logrus.WithField("deck_id", deck.ID).Debugf("Deck updated, new card count=%v", len(deck.Cards))

	response := DrawResponse{Cards: drawnCards}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
