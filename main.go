package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
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

	var storage DeckStorage
	storage = NewInMemoryStorage()

	port := os.Getenv("PORT")
	if port == "" {
		logrus.Fatal("PORT environment variable is not set")
	}

	http.HandleFunc("/decks/", func(w http.ResponseWriter, r *http.Request) { handleDeck(storage, w, r) })
	logrus.Infof("Listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
