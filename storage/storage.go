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
	locks map[uuid.UUID]*sync.RWMutex
	mu    sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		decks: make(map[uuid.UUID]deck.Deck),
		locks: make(map[uuid.UUID]*sync.RWMutex),
	}
}

func (s *InMemoryStorage) SaveDeck(ctx context.Context, d deck.Deck) error {
	s.mu.Lock()
	if _, ok := s.locks[d.ID]; !ok {
		s.locks[d.ID] = &sync.RWMutex{}
	}
	s.locks[d.ID].Lock()
	s.decks[d.ID] = d
	s.locks[d.ID].Unlock()
	s.mu.Unlock()
	return nil
}

func (s *InMemoryStorage) GetDeck(ctx context.Context, id uuid.UUID) (deck.Deck, bool) {
	s.mu.Lock()
	lock, ok := s.locks[id]
	if !ok {
		s.mu.Unlock()
		return deck.Deck{}, false
	}
	lock.RLock()
	defer lock.RUnlock()
	s.mu.Unlock()

	d, found := s.decks[id]
	return d, found
}

func (s *InMemoryStorage) DeleteDeck(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	lock, ok := s.locks[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("deck not found")
	}
	lock.Lock()
	delete(s.decks, id)
	delete(s.locks, id)
	lock.Unlock()
	s.mu.Unlock()
	return nil
}

func (s *InMemoryStorage) UpdateDeck(ctx context.Context, d deck.Deck) error {
	s.mu.Lock()
	lock, ok := s.locks[d.ID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("deck with id=%v not found", d.ID)
	}

	lock.Lock()
	defer lock.Unlock()
	s.mu.Unlock()

	if _, found := s.decks[d.ID]; !found {
		return fmt.Errorf("deck with id=%v not found", d.ID)
	}
	s.decks[d.ID] = d
	return nil
}
