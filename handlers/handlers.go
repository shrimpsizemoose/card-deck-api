package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"deck-of-cards/deck"
	"deck-of-cards/storage"
)

type DeckResponse struct {
	DeckID    string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
}

type OpenDeckResponse struct {
	DeckID    string      `json:"deck_id"`
	Shuffled  bool        `json:"shuffled"`
	Remaining int         `json:"remaining"`
	Cards     []deck.Card `json:"cards"`
}

type DrawResponse struct {
	Cards []deck.Card `json:"cards"`
}

type Handler struct {
	st      storage.DeckStorage
	uuidGen func() uuid.UUID
}

func NewHandler(st storage.DeckStorage) *Handler {
	return &Handler{
		st: st,
		uuidGen: func() uuid.UUID {
			return uuid.New()
		},
	}
}

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

// creates the deck and saves it in the DeckStorage
func (h *Handler) HandleCreateDeck(w http.ResponseWriter, r *http.Request) {
	log := logrus.WithFields(logrus.Fields{"endpoint": "handleCreateDeck"})
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	shuffle := r.URL.Query().Get("shuffle") == "true"
	cardsParam := r.URL.Query().Get("cards")
	cardCodes := parseCardCodes(cardsParam)
	log.Debugf("Request to create a new deck shuffle=%v cards=%v", shuffle, cardsParam)

	id := h.uuidGen()
	d := deck.NewDeck(id, shuffle, cardCodes)
	err := h.st.SaveDeck(r.Context(), *d)
	if err != nil {
		http.Error(w, "Error saving created deck", http.StatusInternalServerError)
	}
	log.WithField("deck_id", d.ID).Debugf("Saving new deck")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/decks/"+d.ID.String())
	w.WriteHeader(http.StatusCreated)

	response := DeckResponse{
		DeckID:    d.ID.String(),
		Shuffled:  d.Shuffled,
		Remaining: len(d.Cards),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fetches the deck from the DeckStore and opens it
func (h *Handler) HandleOpenDeck(w http.ResponseWriter, r *http.Request) {
	deckIDParam := r.PathValue("id")
	log := logrus.WithFields(logrus.Fields{
		"endpoint": "handleOpenDeck",
		"deck_id":  deckIDParam,
	})
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

	d, found := h.st.GetDeck(r.Context(), deckID)
	if !found {
		http.Error(w, fmt.Sprintf("Deck not found: %s", err), http.StatusNotFound)
		return
	}

	log.Debug("Opening deck")

	response := OpenDeckResponse{
		DeckID:    deckID.String(),
		Shuffled:  d.Shuffled,
		Remaining: len(d.Cards),
		Cards:     d.Cards,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fetches the deck from the DeckStorage, draws cards, updates deck
func (h *Handler) HandleDrawCards(w http.ResponseWriter, r *http.Request) {
	deckIDParam := r.PathValue("id")
	log := logrus.WithFields(logrus.Fields{
		"endpoint": "handleDrawCards",
		"deck_id":  deckIDParam,
	})
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

	ctx := r.Context()
	d, found := h.st.GetDeck(ctx, deckID)
	if !found {
		http.Error(w, "Deck not found", http.StatusNotFound)
		return
	}

	numCardsParam := r.URL.Query().Get("count")
	numCards, err := strconv.Atoi(numCardsParam)
	if err != nil || numCards < 1 {
		http.Error(w, "Invalid number of cards", http.StatusBadRequest)
		return
	}
	if numCards > len(d.Cards) {
		http.Error(w, "Not enough cards in the deck", http.StatusBadRequest)
		return
	}

	log.Debugf("Drawing count=%v cards from deck", numCards)
	drawnCards := d.Draw(numCards)
	err = h.st.UpdateDeck(ctx, d)
	if err != nil {
		http.Error(w, "Error updating deck in storage", http.StatusInternalServerError)
	}
	log.Debugf("Deck updated, new card count=%v", len(d.Cards))

	response := DrawResponse{Cards: drawnCards}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
