package main

import (
	"math/rand"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Card struct {
	Value string `json:"value"`
	Suit  string `json:"suit"`
	Code  string `json:"code"`
}

type Deck struct {
	ID       uuid.UUID `json:"deck_id"`
	Shuffled bool      `json:"shuffled"`
	Cards    []Card    `json:"cards"`
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
	d.Shuffled = true
}

// the case when more cards were requested is handled in http handler
func (d *Deck) Draw(numCards int) []Card {
	if numCards > len(d.Cards) {
		numCards = len(d.Cards)
	}
	drawn := d.Cards[:numCards]
	d.Cards = d.Cards[numCards:]
	return drawn
}

func NewDeck(id uuid.UUID, shuffle bool, cardCodes []string) *Deck {
	var cards []Card
	if len(cardCodes) > 0 {
		cards = generateDeckFromCodes(cardCodes)
	} else {
		cards = generateFullDeck()
	}
	deck := &Deck{
		ID:       id,
		Cards:    cards,
		Shuffled: shuffle,
	}
	if shuffle {
		deck.Shuffle()
	}
	return deck
}

func generateFullDeck() []Card {
	suits := []string{"Spades", "Clubs", "Diamonds", "Hearts"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	var cards []Card

	for _, suit := range suits {
		for _, value := range values {
			cards = append(cards, Card{
				Value: value,
				Suit:  suit,
				Code:  value[:1] + suit[:1],
			})
		}
	}

	logrus.Debugf("Number of cards generated: %d", len(cards))
	return cards
}

func generateDeckFromCodes(codes []string) []Card {
	fullDeck := generateFullDeck()
	var cards []Card

	for _, code := range codes {
		for _, card := range fullDeck {
			if card.Code == code {
				cards = append(cards, card)
				break
			}
		}
	}

	logrus.Debugf("Number of cards generated: %d", len(cards))
	return cards
}
