package storage

import (
	"context"
	"reflect"
	"sync"

	"testing"

	"github.com/google/uuid"

	"deck-of-cards/deck"
)

func TestNewInMemoryStorage(t *testing.T) {
	st := NewInMemoryStorage()

	var wg sync.WaitGroup

	count := 10
	wg.Add(count)

	for i := 0; i < count; i++ {
		ctx := context.Background()
		go func() {
			d := deck.NewDeck(uuid.New(), false, nil)
			if err := st.SaveDeck(ctx, *d); err != nil {
				t.Errorf("Error saving deck")
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func TestSaveAndGetDeck(t *testing.T) {
	s := NewInMemoryStorage()
	id := uuid.New()
	d := deck.NewDeck(id, false, nil)
	ctx := context.Background()
	err := s.SaveDeck(ctx, *d)
	if err != nil {
		t.Errorf("SaveDeck failed: %s", err)
	}
	dd, found := s.GetDeck(ctx, id)
	if !found {
		t.Errorf("Deck was not found after creation")
	}

	// I would use probably some external package to make it look less
	if !reflect.DeepEqual(d, &dd) {
		t.Errorf("Saved deck and retrieved deck are not the same")
	}
}

// I don't use this method in service but I have it in the deck, so better test it, eh?
func TestDeleteDeck(t *testing.T) {
	s := NewInMemoryStorage()
	id := uuid.New()
	d := deck.NewDeck(id, false, nil)
	ctx := context.Background()
	if err := s.SaveDeck(ctx, *d); err != nil {
		t.Errorf("Error saving deck")
	}
	if err := s.DeleteDeck(ctx, d.ID); err != nil {
		t.Errorf("Error deleting deck")
	}
	_, found := s.GetDeck(ctx, d.ID)
	if found {
		t.Errorf("Deck was found after deletion")
	}
}

func TestUpdateDeck(t *testing.T) {
	s := NewInMemoryStorage()
	id := uuid.New()

	d := deck.NewDeck(id, false, nil)
	ctx := context.Background()
	_ = s.SaveDeck(ctx, *d)

	d.Shuffle()
	err := s.UpdateDeck(ctx, *d)
	if err != nil {
		t.Errorf("UpdateDeck failed: %s", err)
	}

	dd, found := s.GetDeck(ctx, d.ID)
	if !found {
		t.Errorf("Somehow, updated deck not found")
	}

	if !reflect.DeepEqual(d, &dd) {
		t.Errorf("Updated deck does not match")
	}
}
