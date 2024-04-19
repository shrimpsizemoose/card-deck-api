package main

import (
	"fmt"

	"github.com/google/uuid"
)

type DeckStorage interface {
	SaveDeck(deck Deck) error
	GetDeck(id uuid.UUID) (Deck, error)
	DeleteDeck(id uuid.UUID) error
	UpdateDeck(deck Deck) error
}

type InMemoryStorage struct {
	decks map[uuid.UUID]Deck
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		decks: make(map[uuid.UUID]Deck),
	}
}

func (s *InMemoryStorage) SaveDeck(deck Deck) error {
	s.decks[deck.ID] = deck
	return nil
}

func (s *InMemoryStorage) GetDeck(id uuid.UUID) (Deck, error) {
	deck, found := s.decks[id]
	if !found {
		return Deck{}, fmt.Errorf("deck not found: id=%v is not in InMemoryStorage", id)
	}
	return deck, nil
}

func (s *InMemoryStorage) DeleteDeck(id uuid.UUID) error {
	delete(s.decks, id)
	return nil
}

func (s *InMemoryStorage) UpdateDeck(deck Deck) error {
	_, found := s.decks[deck.ID]
	if !found {
		return fmt.Errorf("deck with id=%v not found", deck.ID)
	}
	s.decks[deck.ID] = deck
	return nil
}
