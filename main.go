package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"deck-of-cards/handlers"
	"deck-of-cards/storage"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("Logger initialized")
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	debug := os.Getenv("DEBUG")
	if debug == "1" {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Logging debug output, run without DEBUG=1 to disable")
	}

	var st storage.DeckStorage
	st = storage.NewInMemoryStorage()

	h := handlers.NewHandler(st)

	port := os.Getenv("PORT")
	if port == "" {
		logrus.Fatal("PORT environment variable is not set")
	}

	http.HandleFunc("POST /decks/", h.HandleCreateDeck)
	http.HandleFunc("GET /decks/{id}", h.HandleOpenDeck)
	http.HandleFunc("POST /decks/{id}/draw", h.HandleDrawCards)

	logrus.Infof("Listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
