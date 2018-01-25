package api

import (
	"github.com/go-chi/chi"
	"github.com/maddevsio/comedian/storage"
	"log"
	"net/http"
	//"github.com/sirupsen/logrus"
	"fmt"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
)

type (
	// API struct
	API struct {
		db *storage.Storage
	}
)

func GetHandler(conf config.Config) (http.Handler, error) {
	db, err := storage.NewMySQL(conf)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	router.Post("/commands", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		fmt.Printf("%+v\n", r.Form)
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
				err = addComedian(db, text)
				if err != nil {
					log.Printf("Failed to add standup_user: %s\n", err)
					http.Error(w, "Failed to add a comedian!", 500)
					return
				}
				w.Write([]byte("Comedian added!"))
			case "/comedianremove":
				//err = RemoveComedian(text)
				if err != nil {
					http.Error(w, "Failed to remove a comedian!", 500)
					return
				}
				w.Write([]byte("Comedian removed!"))
			case "/comedianlist":
				//list, err := GetComedianList()
				if err != nil {
					http.Error(w, "Failed to list comedians!", 500)
					return
				}
				//w.Write([]byte(list))
			}
		} else {
			http.Error(w, "No command!", 400)
		}
	})
	return router, nil
}

func addComedian(database *storage.MySQL, text string) error {
	log.Printf("Adding comedian: %s\n", text)
	user := model.StandupUser{
		SlackName: text,
	}
	_, err := database.CreateStandupUser(user)
	return err
}
