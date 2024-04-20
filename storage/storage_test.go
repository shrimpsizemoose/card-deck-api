package storage

import (
	"context"
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
			st.SaveDeck(ctx, *d)

			wg.Done()
		}()
	}

	wg.Wait()
}
