package api

import (
	"github.com/go-chi/chi"
	"github.com/maddevsio/comedian/storage"
	"net/http"
	"log"
	"fmt"
)

type (
	// API struct
	API struct {
		db *storage.Storage
	}
)

var comedians = make([]string, 0)

func GetHandler() http.Handler {
	router := chi.NewRouter()

	router.Post("/commands", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid request!", 400)
			return
		}
		if commands, ok := r.Form["command"]; ok && len(commands) > 0 {
			command := r.Form["command"][0]
			text := ""
			if texts, ok := r.Form["text"]; ok && len(texts) > 0 {
				text = r.Form["text"][0]
			}
			switch command {
			case "/comedianadd":
				err = AddComedian(text)
				if err != nil {
					http.Error(w, "Failed to add a comedian!", 500)
				}
				w.Write([]byte("Comedian added!"))
			case "/comedianremove":
				err = RemoveComedian(text)
				if err != nil {
					http.Error(w, "Failed to remove a comedian!", 500)
				}
				w.Write([]byte("Comedian removed!"))
			case "/comedianlist":
				list, err := GetComedianList()
				if err != nil {
					http.Error(w, "Failed to list comedians!", 500)
				}
				w.Write([]byte(list))
			}
		} else {
			http.Error(w, "No command!", 400)
		}
	})
	return router
}

func AddComedian(text string) error {
	log.Printf("Adding comedian: %s\n", text)
	comedians = append(comedians, text)
	return nil
}

func RemoveComedian(text string) error {
	log.Printf("Removing comedian: %s\n", text)
	for i, c := range comedians {
		if c == text {
			comedians = append(comedians[:i], comedians[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no comedian \"%s\" found", text)
}

func GetComedianList() (string, error) {
	list := "Comedians:\n"
	for i, c := range comedians {
		list += fmt.Sprintf("%d: %s\n", i, c)
	}
	return list, nil
}