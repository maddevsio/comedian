package api

import (
	"github.com/go-chi/chi"
	"github.com/maddevsio/comedian/storage"
	"net/http"
)

type (
	// API struct
	API struct {
		db *storage.Storage
	}
)

func GetHandler() http.Handler {
	router := chi.NewRouter()

	router.Post("/commands", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HI!"))
	})
	return router
}
