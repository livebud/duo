package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/livebud/duo"
	"github.com/livebud/duo/example/hn/controller"
	"github.com/livebud/mux"
	"github.com/matthewmueller/hackernews"
)

func main() {
	log := slog.Default()
	hn := hackernews.New()
	view := duo.New(os.DirFS("view"))
	controller := controller.New(hn, view)
	router := mux.New()
	router.Get("/", controller.Index)
	router.Get("/{id}", controller.Show)
	log.Info("listening on http://localhost:3002")
	if err := http.ListenAndServe(":3002", router); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
