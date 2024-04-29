# Card Deck T**** take-home assignment

Implementation of REST API service to simulate a deck of cards.

**This code requires Go 1.22** because it uses [routing enhancement](https://go.dev/blog/routing-enhancements)

Building the service and running in docker locally:

```bash
make @docker-build
make @docker-run
```

Makefile is set to serve on port 8088. The implementation uses in-memory storage and all decks would be lost on restart. You can use additional make targets to interact with the service, like this:

```bash
make local-http-create-shuffled-deck  # creates a deck and returns deck_id
DECK_ID=66960c43-c727-4e33-82da-ede4f204f484 make local-http-open-deck-jq
DECK_ID=66960c43-c727-4e33-82da-ede4f204f484 make local-http-draw-deck-jq
```

Note that `DECK_ID` in the last two commands is the one returned by the first one. If you don't have jq installed, the last two commands can be used as `local-http-open-deck` and `local-http-draw-deck` correspondingly

## Card deck

All cards are assumed to be from the deck of standard French 52-card deck. It includes thirteen ranks in four suits: (♣), diamonds (♦), hearts (♥), and spades (♠). Joker cards are not supported.

For simplicity, all cards are coded with a 2-3 characters string, like "KD" for "King of Diamonds", "AS" for "Ace of Spades", "10C" for "Ten of Clubs", and so on. The user can provide a subset of card codes when creating a card, but unknown codes will be ignored.

The service does assumes anything about the deck you want to create, that is if you want a deck consisting of 20 Aces of hearts, the service would be happy to create it.

## Endpoints

API provides parameters to Create card decks and Open and Draw cards from existing decks.

### Creating a Deck `POST /decks/`

**NB the end slash** in path.

Creates a Deck of cards. By default, all 52 cards are used and the deck is created sequentially using standard [Preferans](https://en.wikipedia.org/wiki/Preferans) progression: ♠<♣<♦<♥, that is "AS, 2S, 3S, ...QS, KS, AC, 2C, 3C, ...". 

**URL Parameters for `POST /decks/`**

| Parameter | Required | Description                                              |
| --------- | -------- | -------------------------------------------------------- |
| shuffle   | no       | whether the deck should be shuffled on creation          |
| cards     | no       | optional list of card keys to use when creating the deck |

When no parameters are provided, returns a deck consisting of 52 cards in sequential order. There's no duplication checks on the cards provides, but the card codes not in the deck would be ignored

#### Example Success Response from `POST /decks/`

**Code:** 201 CREATED

```json
{
    "deck_id": "e13aaa48-2f62-4457-8c87-790cd856d536",
    "shuffled": "false",
    "remaining": 52,
}
```

#### Example Success Response from `POST /decks/?shuffle=true,cards=AS,AS,AS,KH,KD,GG,IDDQD`

**Code:** 201 CREATED

```json
{
    "deck_id": "118a1a98-2fd2-44d9-83d2-b34fe4bd5230",
    "shuffled": "true",
    "remaining": 5,
}
```

The resulting deck is stored in DeckStorage and can be accessed using the returned ID

### Open Deck `GET /decks/{uuid}`

This opens the deck given the Deck ID and returns deck properties and cards. The deck_id is provided as a path parameter, for example

```text
GET /decks/e13aaa48-2f62-4457-8c87-790cd856d536
```

If the deck ID is wrong or the deck is not found, it would error with some status code (400, 404, 405) depending on the situation

#### Example Success Response from `GET /decks/{uuid}`

**Code:** 200 OK

```json
{
  "deck_id": "b63feb43-cd9a-4376-8560-84082569e736",
  "shuffled": false,
  "remaining": 2,
  "cards": [
    {
      "value": "QUEEN",
      "suit": "Hearts",
      "code": "QH"
    },
    {
      "value": "KING",
      "suit": "Hearts",
      "code": "KH"
    },
    {
      "value": "10",
      "suit": "Hearts",
      "code": "10H"
    }
  ]
}
```

This request does not alter the deck and can be used for RO access to the deck

### Draw a card from Deck `POST /decks/{uuid}/draw?count=N`

This opens the deck given the Deck ID and returns deck properties along with its cards

**URL Parameters for `POST /decks/{uuid}/draw?count=N`**

| Parameter | Required | Description                                             |
| --------- | -------- | ------------------------------------------------------- |
| count     | yes      | amount of cards to draw from the deck. Should be integer |

The deck_id is provided as a path parameter, for example

```text
POST /decks/e13aaa48-2f62-4457-8c87-790cd856d536/draw?count=5
```

If the deck ID is wrong or the deck is not found, or something else is wrong, it would error with some status code (400, 404, 405) depending on the situation

#### Example Success Response for `POST /decks/{uuid}/draw?count=N`

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
      "code": "10S"
    },
    {
      "value": "7",
      "suit": "Diamonds",
      "code": "7D"
    }
  ]
}
```

This request updates the deck: after the draw, the deck would contain `count` fewer cards.

## Buliding

Local build builds the executable for the service which can be run as `./card-deck-api`:

```bash
make @build
```

Building with docker:

```bash
make @docker-build
```

### Running locally

```bash
make @run
```

You can use the provided Makefile file to change PORT. Set env variable DEBUG=1 to enable debug logging or use `make local-debug-run`

### Building and running in Docker locally

```bash
make @docker-build
make @run-docker
```

## TODO (What would I do if I had more time)

* I'm quite fond of Golang standard library, but I think some lib for marshaling would be helpful here, so I'd probably use _whatever the team is using_ or gin/negroni/chi
* The handlers package looks very verbose with all that validation and type coercion, this would be very annoyng to work with when adding new handlers, so it would require refactoring for better flexibility in the future
* Ratelimit would be nice if the service expected to handle some high traffic (let's say 100+ RPS), and there should be a way to manage storage timeouts
* Once there's ratelimit, then there might be reasons to add things like auth and such to figure out per-account quotas
* More tests would be always nice to have, especially if external storage is used
* Storage should happen in an external system, which would help with state management and can help with consistency
* For external storage one should use [singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight) to help with parallel requests
* It would be annoying to add new cards or change deck types, which would require some refactoring if such a thing is needed. Having a smaller deck would work fine with the current service, but adding different cards for example for [mus](https://en.wikipedia.org/wiki/Mus_(card_game)) would be challenging
* Would probably use more context handling, adding timeouts and such. I've added it after once I made the storage package
* I wanted to use stdlib as much as possible with the exception of logrus, but for "real-world" logging I would probably use [uber-go/zap](https://github.com/uber-go/zap) instead of logrus. I think logrus is more commonly used though (maybe?)
* Again, this is more like my own implied limitation of writing a lean service with little amout of external libs, but currently there's no monitoring and likely in production it should have prometheus handler installed

### Extending storage

For example, adding Redis would be something like

```go
type RedisStorage struct {
    client *redis.Client
}

func (s *RedisStorage) SaveDeck(ctx context.Context, d deck.Deck) error {
    // probably saving deck-id as deck:uuid as hashmap and cards:uuid as list
}

func (s *RedisStorage) GetDeck(ctx context.Context, id uuid.UUID) (Deck, bool) {
    // should get cards from redis
}

// etc
```

### Adding new handlers

Add handlers to the [handlers.go](./handlers/handlers.go) file, and register in [main.go](./main.go), and any deck logic should go into [deck.go](./deck/deck.go)
