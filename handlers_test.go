package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	// "strings"
	"testing"

	"github.com/google/uuid"
)

var (
	fakeUUID = uuid.MustParse("0000aaaa-0000-0000-0000-0000aaaa0000")
	storage  = NewInMemoryStorage()
)

func TestHandleCreateDeck5CardNoShufflingSurprises(t *testing.T) {
	origUUIDFunc := generateUUID
	generateUUID = func() uuid.UUID {
		return fakeUUID
	}
	defer func() { generateUUID = origUUIDFunc }()

	// Setup a request to pass to our handler.
	cards := []string{"AS", "KD", "AC", "2C", "KH"}
	req, err := http.NewRequest("POST", "/decks/?cards=AS,KD,AC,2C,KH&shuffle=false", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleCreateDeck(storage, w, r) })

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp DeckResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	if resp.DeckID != fakeUUID.String() {
		t.Errorf("handler returned unexpected deck_id: got %v want %v", resp.DeckID, fakeUUID.String())
	}

	if resp.Shuffled {
		t.Errorf("handler returned unexpected shuffle status: got %v want %v", resp.Shuffled, false)
	}

	if resp.Remaining != 5 {
		t.Errorf("handler returned unexpected remaining count: got %v want %v", resp.Remaining, 5)
	}

	deck, err := storage.GetDeck(fakeUUID)
	if err != nil {
		t.Errorf("Service reported deck created, but error fetching it from storage")
	}

	for i, code := range cards {
		if code != deck.Cards[i].Code {
			t.Errorf("Creting deck of cards differs in position %d. Expected %s got %s", i, code, deck.Cards[i].Code)
		}
	}

}

func TestHandleDeckCreate(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		expectedStatus   int
		expectedShuffled bool
		expectedNumCards int
	}{
		{"Can create shuffled deck of 5 cards", "POST", "/decks/?shuffle=true&cards=AS,KD,QH,2C,3S", http.StatusCreated, true, 5},
		{"Can create unshuffled deck of 4 cards", "POST", "/decks/?shuffle=false&cards=AS,KD,QH,2C", http.StatusCreated, false, 4},
		{"Can create implicitly unshuffled deck of 3 cards", "POST", "/decks/?cards=AS,KD,QH", http.StatusCreated, false, 3},
		{"Can create shuffled deck of 52 cards", "POST", "/decks/?shuffle=true", http.StatusCreated, true, 52},
		{"Defaults creates deck of 52 cards", "POST", "/decks/", http.StatusCreated, false, 52},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := bytes.NewBuffer(nil)
			req, _ := http.NewRequest(tc.method, tc.path, requestBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleDeck(storage, w, r) })

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusCreated {
				var resp DeckResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Shuffled != tc.expectedShuffled {
					t.Errorf("service reports different shuffle stated, expected %v, got %v", tc.expectedShuffled, resp.Shuffled)
				}
				if resp.Remaining != tc.expectedNumCards {
					t.Errorf("service reports wrong number of cards, expected %d, got %d", tc.expectedNumCards, resp.Remaining)
				}

				id, err := uuid.Parse(resp.DeckID)
				if err != nil {
					t.Errorf("Cannot parse returned UUID when creating deck")
				}
				deck, err := storage.GetDeck(id)
				if err != nil {
					t.Errorf("Service reported deck created, but error fetching it from storage")
				}

				if deck.Shuffled != tc.expectedShuffled {
					t.Errorf("Storage reports wrong shuffle state, expected %v, got %v", tc.expectedShuffled, resp.Shuffled)
				}
				if len(deck.Cards) != tc.expectedNumCards {
					t.Errorf("storage reports wrong number of cards, expected %d, got %d", tc.expectedNumCards, len(deck.Cards))
				}
			}
		})
	}
}

func TestHandleDrawCards(t *testing.T) {
	mock := NewDeck(fakeUUID, false, []string{"AS", "KD", "QH", "2C"})
	storage.SaveDeck(*mock)

	tests := []struct {
		name           string
		method         string
		deckID         string
		numCards       string
		expectedStatus int
	}{
		{"Method Not Allowed", "GET", mock.ID.String(), "1", http.StatusMethodNotAllowed},
		{"Missing Deck ID", "POST", "", "1", http.StatusBadRequest},
		{"Invalid Deck ID", "POST", "invalid-uuid", "1", http.StatusBadRequest},
		{"Deck Not Found", "POST", uuid.New().String(), "1", http.StatusNotFound},
		{"Invalid Number of Cards", "POST", mock.ID.String(), "abc", http.StatusBadRequest},
		{"Draw Zero Cards", "POST", mock.ID.String(), "0", http.StatusBadRequest},
		{"More Cards Than Available", "POST", mock.ID.String(), "10", http.StatusBadRequest},
		{"Successful Draw", "POST", mock.ID.String(), "2", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := bytes.NewBuffer(nil)
			req, _ := http.NewRequest(tc.method, "/decks/"+tc.deckID+"/draw?count="+tc.numCards, requestBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleDeck(storage, w, r) })

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response DrawResponse
				json.NewDecoder(rr.Body).Decode(&response)
				if len(response.Cards) != 2 {
					t.Errorf("expected 2 cards, got %d", len(response.Cards))
				}
			}
		})
	}
}
