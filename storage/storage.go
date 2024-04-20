package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"deck-of-cards/deck"
)

type DeckStorage interface {
	SaveDeck(ctx context.Context, d deck.Deck) error
	GetDeck(ctx context.Context, id uuid.UUID) (deck.Deck, error)
	DeleteDeck(ctx context.Context, id uuid.UUID) error
	UpdateDeck(ctx context.Context, d deck.Deck) error
}

type InMemoryStorage struct {
	decks map[uuid.UUID]deck.Deck
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		decks: make(map[uuid.UUID]deck.Deck),
	}
}

func (s *InMemoryStorage) SaveDeck(ctx context.Context, d deck.Deck) error {
	s.decks[d.ID] = d
	return nil
}

func (s *InMemoryStorage) GetDeck(ctx context.Context, id uuid.UUID) (deck.Deck, error) {
	d, found := s.decks[id]
	if !found {
		return deck.Deck{}, fmt.Errorf("deck not found: id=%v is not in InMemoryStorage", id)
	}
	return d, nil
}

func (s *InMemoryStorage) DeleteDeck(ctx context.Context, id uuid.UUID) error {
	delete(s.decks, id)
	return nil
}

func (s *InMemoryStorage) UpdateDeck(ctx context.Context, d deck.Deck) error {
	_, found := s.decks[d.ID]
	if !found {
		return fmt.Errorf("deck with id=%v not found", d.ID)
	}
	s.decks[d.ID] = d
	return nil
}
