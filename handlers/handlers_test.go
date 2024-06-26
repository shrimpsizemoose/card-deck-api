package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"net/http/httptest"
	"testing"

	"deck-of-cards/deck"
	"deck-of-cards/storage"

	"github.com/google/uuid"
)

var (
	fakeUUID = uuid.MustParse("0000aaaa-0000-0000-0000-0000aaaa0000")
)

func TestParseCardCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected number of card codes
	}{
		{"Empty string", "", 0},
		{"Single card code", "AS", 1},
		{"Multiple card codes", "AS,KD,5H", 3},
		{"With space", "AS, KD, 5H", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCardCodes(tt.input)
			if len(got) != tt.want {
				t.Errorf("parseCardCodes(%q) = %v, want %v", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestParseCardCodesTrimSpaces(t *testing.T) {
	input := "AS, KD, 5H"
	want := []string{"AS", "KD", "5H"}
	got := parseCardCodes(input)

	if len(got) != len(want) {
		t.Fatalf("Expected slice length %d, got %d", len(want), len(got))
	}

	for i, w := range want {
		if got[i] != w {
			t.Errorf("At index %d, Expected %s, got %s", i, w, got[i])
		}
	}
}

func TestHandleCreateDeck5CardNoShufflingSurprises(t *testing.T) {
	h := &Handler{
		st: storage.NewInMemoryStorage(),
		uuidGen: func() uuid.UUID {
			return fakeUUID
		},
	}

	// Setup a request to pass to our handler.
	cards := []string{"AS", "KD", "AC", "2C", "KH"}
	req, err := http.NewRequest("POST", "/decks/?cards=AS,KD,AC,2C,KH&shuffle=false", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.HandleCreateDeck)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response DeckResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	if response.DeckID != fakeUUID.String() {
		t.Errorf("handler returned unexpected deck_id: got %v want %v", response.DeckID, fakeUUID.String())
	}

	if response.Shuffled {
		t.Errorf("handler returned unexpected shuffle status: got %v want %v", response.Shuffled, false)
	}

	if response.Remaining != 5 {
		t.Errorf("handler returned unexpected remaining count: got %v want %v", response.Remaining, 5)
	}

	d, found := h.st.GetDeck(req.Context(), fakeUUID)
	if !found {
		t.Errorf("Service reported deck created, but it seems to be missing from storage")
	}

	for i, code := range cards {
		if code != d.Cards[i].Code {
			t.Errorf("Creating deck of cards differs in position %d. Expected %s got %s", i, code, d.Cards[i].Code)
		}
	}

}

func TestHandleDeckCreate(t *testing.T) {
	h := NewHandler(storage.NewInMemoryStorage())

	tests := []struct {
		name             string
		method           string
		path             string
		expectedStatus   int
		expectedShuffled bool
		expectedNumCards int
	}{
		{"Method Not Allowed", "GET", "/decks/", http.StatusMethodNotAllowed, false, 0},
		{"Can create shuffled deck of 5 cards", "POST", "/decks/?shuffle=true&cards=AS,KD,QH,2C,3S", http.StatusCreated, true, 5},
		{"Can create unshuffled deck of 4 cards", "POST", "/decks/?shuffle=false&cards=AS,KD,QH,2C", http.StatusCreated, false, 4},
		{"Can create implicitly unshuffled deck of 3 cards", "POST", "/decks/?cards=AS,KD,QH", http.StatusCreated, false, 3},
		{"Can create shuffled deck of 52 cards", "POST", "/decks/?shuffle=true", http.StatusCreated, true, 52},
		{"Can create unshuffled deck of 10 same cards", "POST", "/decks/?cards=AS,AS,AS,AS,AS,AS,AS,AS,AS,AS", http.StatusCreated, false, 10},
		{"Unknown card codes are ignored", "POST", "/decks/?cards=AS,AS,AS,KH,KD,GG,IDDQD", http.StatusCreated, false, 5},
		{"Defaults creates deck of 52 cards", "POST", "/decks/", http.StatusCreated, false, 52},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := bytes.NewBuffer(nil)
			req, _ := http.NewRequest(tc.method, tc.path, requestBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.HandleCreateDeck)

			handler.ServeHTTP(rr, req)
			ctx, cancel := context.WithTimeout(req.Context(), time.Second)
			defer cancel()

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusCreated {
				var resp DeckResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Error("Error decoding server response")
				}
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
				deck, found := h.st.GetDeck(ctx, id)
				if !found {
					t.Errorf("Service reported deck created, but it seems to be missing from storage")
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

func TestHandleOpenDeck(t *testing.T) {

	tests := []struct {
		name           string
		method         string
		deckID         string
		expectedStatus int
	}{
		{"Method Not Allowed", "POST", fakeUUID.String(), http.StatusMethodNotAllowed},
		{"Missing Deck ID", "GET", "", http.StatusBadRequest},
		{"Invalid Deck ID", "GET", "invalid-uuid", http.StatusBadRequest},
		{"Deck Not Found", "GET", uuid.New().String(), http.StatusNotFound},
		{"Successful Open", "GET", fakeUUID.String(), http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			h := NewHandler(storage.NewInMemoryStorage())
			mock := deck.NewDeck(fakeUUID, false, []string{"AS", "KD", "QH", "2C"})
			if err := h.st.SaveDeck(ctx, *mock); err != nil {
				t.Errorf("Error saving dummiy deck in storage")
			}
			expectedShuffled := mock.Shuffled
			expectedRemaining := len(mock.Cards)

			requestBody := bytes.NewBuffer(nil)
			req, _ := http.NewRequestWithContext(ctx, tc.method, "/decks/"+tc.deckID, requestBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.HandleOpenDeck)
			req.SetPathValue("id", tc.deckID)

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response OpenDeckResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Error("Error decoding server response")
				}
				if response.DeckID != tc.deckID {
					t.Errorf("got back wrong DeckID somehow, expected %s, got %s", tc.deckID, response.DeckID)
				}
				if response.Remaining != expectedRemaining {
					t.Errorf("expected %d cards, got %d", expectedRemaining, response.Remaining)
				}
				if response.Shuffled != expectedShuffled {
					t.Errorf("shuffled state mismatch: expected %v, got %v", expectedShuffled, response.Shuffled)
				}
			}
		})
	}
}

func TestHandleDrawCards(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		deckID           string
		numCards         string
		expectedStatus   int
		expectedNumCards int
	}{
		{"Method Not Allowed", "GET", fakeUUID.String(), "1", http.StatusMethodNotAllowed, 0},
		{"Missing Deck ID", "POST", "", "1", http.StatusBadRequest, 0},
		{"Invalid Deck ID", "POST", "invalid-uuid", "1", http.StatusBadRequest, 0},
		{"Deck Not Found", "POST", uuid.New().String(), "1", http.StatusNotFound, 0},
		{"Invalid Number of Cards", "POST", fakeUUID.String(), "abc", http.StatusBadRequest, 0},
		{"Draw Zero Cards", "POST", fakeUUID.String(), "0", http.StatusBadRequest, 0},
		{"More Cards Than Available", "POST", fakeUUID.String(), "10", http.StatusBadRequest, 0},
		{"Successful Draw 3 Cards", "POST", fakeUUID.String(), "3", http.StatusOK, 3},
		{"Successful Draw 2 Cards", "POST", fakeUUID.String(), "2", http.StatusOK, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			h := NewHandler(storage.NewInMemoryStorage())
			mock := deck.NewDeck(fakeUUID, false, []string{"AS", "KD", "QH", "2C"})
			if err := h.st.SaveDeck(ctx, *mock); err != nil {
				t.Error("Error saving dummy deck in storage")
			}

			requestBody := bytes.NewBuffer(nil)
			req, _ := http.NewRequestWithContext(ctx, tc.method, "/decks/"+tc.deckID+"/draw?count="+tc.numCards, requestBody)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.HandleDrawCards)
			req.SetPathValue("id", tc.deckID)

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response DrawResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Error("Error decoding server response")
				}
				if len(response.Cards) != tc.expectedNumCards {
					t.Errorf("expected %d cards, got %d", tc.expectedNumCards, len(response.Cards))
				}
			}
		})
	}
}
