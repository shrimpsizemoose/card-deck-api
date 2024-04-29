package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"deck-of-cards/deck"
)

type DeckStorage interface {
	SaveDeck(ctx context.Context, d deck.Deck) error
	GetDeck(ctx context.Context, id uuid.UUID) (deck.Deck, bool)
	DeleteDeck(ctx context.Context, id uuid.UUID) error
	UpdateDeck(ctx context.Context, d deck.Deck) error
}

type InMemoryStorage struct {
	decks map[uuid.UUID]deck.Deck
	mu    sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		decks: make(map[uuid.UUID]deck.Deck),
	}
}

func (s *InMemoryStorage) SaveDeck(ctx context.Context, d deck.Deck) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.decks[d.ID] = d
	return nil
}

func (s *InMemoryStorage) GetDeck(ctx context.Context, id uuid.UUID) (deck.Deck, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, found := s.decks[id]
	return d, found
}

func (s *InMemoryStorage) DeleteDeck(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.decks, id)
	return nil
}

func (s *InMemoryStorage) UpdateDeck(ctx context.Context, d deck.Deck) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.decks[d.ID]; !found {
		return fmt.Errorf("deck with id=%v not found", d.ID)
	}
	s.decks[d.ID] = d
	return nil
}
