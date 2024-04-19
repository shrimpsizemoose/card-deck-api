package main

import (
	"github.com/google/uuid"
	"testing"
)

func TestNewDeckFullDeck(t *testing.T) {
	deck := NewDeck(uuid.New(), false, nil)
	if len(deck.Cards) != 52 {
		t.Errorf("Expected 52 cards in a new deck, got %d", len(deck.Cards))
	}
}

func TestNewDeckWithSpecificCards(t *testing.T) {
	cards := []string{"AS", "KD", "AC", "2C", "KH"}
	deck := NewDeck(uuid.New(), false, cards)
	if len(deck.Cards) != 5 {
		t.Errorf("Expected 5 cards in deck, got %d", len(deck.Cards))
	}
	// ugly comparison
	for _, card := range deck.Cards {
		if !(card.Value == "A" && card.Suit == "Spades" || card.Suit == "Clubs") && !(card.Value == "K" && (card.Suit == "Diamonds" || card.Suit == "Hearts")) && !(card.Value == "2" && card.Suit == "Clubs") {
			t.Errorf("Unexpected card %v in deck", card)
		}
	}
}

// can probably get away with just checking a few cards in deck, but checking strings is more thorough
func TestShuffleDeck(t *testing.T) {
	deck := NewDeck(uuid.New(), false, nil)
	if deck.Shuffled {
		t.Errorf("Deck shuffled flag was set while I asked for unshuffled deck")
	}

	deckToString := func(cards []Card) string {
		var result string
		for _, card := range cards {
			result += card.Code
		}
		return result
	}
	preShuffle := deckToString(deck.Cards)

	deck.Shuffle()
	if !deck.Shuffled {
		t.Errorf("Deck shuffled flag not set post shuffle")
	}

	postShuffle := deckToString(deck.Cards)
	if preShuffle == postShuffle {
		t.Errorf("Shuffle did not change the deck order")
	}
}

func TestDrawCards(t *testing.T) {
	deck := NewDeck(uuid.New(), false, nil)
	drawn := deck.Draw(5)
	if len(drawn) != 5 {
		t.Errorf("Expected 5 cards to be drawn, got %d", len(drawn))
	}
	if len(deck.Cards) != 47 {
		t.Errorf("Expected 47 cards remaining in deck, got %d", len(deck.Cards))
	}
}

// I think this should not raise here, but it is handled in handler. This test here to test conformity
func TestDrawMoreCardsThanAvailable(t *testing.T) {
	cards := []string{"AS", "KD"}
	deck := NewDeck(uuid.New(), false, cards)
	drawn := deck.Draw(3)
	if len(drawn) != 2 {
		t.Errorf("Tried to draw more cards than available, expected 2 cards, got %d", len(drawn))
	}
	if len(deck.Cards) != 0 {
		t.Errorf("Expected no cards remaining in deck, got %d", len(deck.Cards))
	}
}

func TestCardValidation(t *testing.T) {
	deck := NewDeck(uuid.New(), false, nil)
	for _, card := range deck.Cards {
		if card.Value == "" || card.Suit == "" || card.Code == "" {
			t.Errorf("Card validation failed, empty attribute found in card: %+v", card)
		}
	}
}
