# Card Deck Toggl take-home assignment

Implementation of REST API service to simulate a deck of cards

Building the service and running in docker locally:
```
make @docker-build
make @docker-run
```

Makefile is set to serve on port 8088. The implementation uses in-memory storage and all decks would be lost on restart.

All cards are assumed to be from the deck of standard French 52-cards deck. It includes thirteen ranks in four suits: (♣), diamonds (♦), hearts (♥) and spades (♠). Jocker cards are not supported.

For simplicity, all cards are coded with 2-3 characters string, like "KD" for "King of Diamonds", "AS" for "Ace of Spades", "10C" for "Ten of Clubs", and so on. When creating a card, user can provide subset of card codes, but unknown codes will be ignored.


## Endpoints

API provides parameters to Create card decks and Open and Draw cards from existing decks.

### Creating a Deck `POST /decks/`

**NB the end slash** in path.

Creates a Deck of cards. By default all 52 cards are used and the deck is created sequentially using standard [Preferans](https://en.wikipedia.org/wiki/Preferans) progression: ♠♣♦♥, that is "AS, 2S, 3S, ...QS, KS, AC, 2C, 3C, ...". 

#### URL Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| shuffle   | no       | whether the deck should be shuffled on creation |
| cards | no | optional list of card keys to use when creating the deck |

When no parameters provided, returns a deck consisting of 52 card in sequential order

#### Example Success Response

**Code:** 200 OK
```json
{
    "deck_id": "e13aaa48-2f62-4457-8c87-790cd856d536",
    "shuffled": "false",
    "remaining": 52,
}
```

Resulting deck is stored in DeckStorage and can be accessed using returned ID

### Open Deck `GET /decks/{uuid}`

This opens the deck given the Deck ID and returns deck properties along with its cards

#### URL Parameters

The deck_id is provided as path parameter, for example
```
GET /decks/e13aaa48-2f62-4457-8c87-790cd856d536
```

If the deck ID is wrong or the deck is not found, it would error with 400/404 depending on the sitation

#### Example Success Response

**Code:** 200 OK
```json
{
  "deck_id": "b63feb43-cd9a-4376-8560-84082569e736",
  "shuffled": false,
  "remaining": 2,
  "cards": [
    {
      "value": "Q",
      "suit": "Hearts",
      "code": "QH"
    },
    {
      "value": "K",
      "suit": "Hearts",
      "code": "KH"
    }
  ]
}
```

This request does not alter the deck in any way

### Draw a card from Deck `POST /decks/{uuid}/draw?count=5`

This opens the deck given the Deck ID and returns deck properties along with its cards

#### URL Parameters

The deck_id is provided as path parameter, for example
```
POST /decks/e13aaa48-2f62-4457-8c87-790cd856d536/draw?count=5
```

If the deck ID is wrong or the deck is not found, it would error with 400/404 depending on the sitation

#### Example Success Response

**Code:** 200 OK
```json
{
  "cards": [
    {
      "value": "4",
      "suit": "Clubs",
      "code": "4C"
    },
    {
      "value": "6",
      "suit": "Diamonds",
      "code": "6D"
    },
    {
      "value": "3",
      "suit": "Clubs",
      "code": "3C"
    },
    {
      "value": "10",
      "suit": "Spades",
      "code": "1S"
    },
    {
      "value": "7",
      "suit": "Diamonds",
      "code": "7D"
    }
  ]
}
```

This request updates the deck: after the draw the deck would contain `count` less cards

## Buliding

Local build builds the executable for the service which can be run as `./card-deck-api`:

```
make @build
```

Building with docker:
```
make @docker-build
```

### Running locally

```
make @run
```

You can use the provided Makefile file to change PORT. Set env variable DEBUG=1 to enable debug output

### Running in Docker

```
make @docker-build
make @run-docker
```

## TODO (What would I do if I had more time)

* I'm quite fond of Golang standard library, but I think some lib for marshmalling would be helpful here, so I'd probably use _whatever the team is using_ or gin/negroni/chi
* If the service expected to handle some high trafic (let's say 100+ RPS), there must be some locking implemented in storage like using `sync.RWMutex` to prevent users consistency issues when doing Draw
* More tests would be always nice to have
* Storage should happen in external system, which potentially would help with consistency as well 
* Right now it's unclear on how to add new cards or change deck types, there's probably a better way to handle that

### Extending storage

For example, adding Redis would be something like

```go
type RedisStorage struct {
    client *redis.Client
}

func (s *RedisStorage) SaveDeck(deck Deck) error {
    // probably saving deck-id as deck:uuid as hashmap and cards:uuid as list
}

func (s *RedisStorage) GetDeck(id uuid.UUID) (Deck, error) {
    // should get cards from redis as well
}

// etc
```

### Adding new handlers

Add handlers to the [handlers.go](./handlers.go) file, and any deck logic should go into [deck.go](./deck.go)
