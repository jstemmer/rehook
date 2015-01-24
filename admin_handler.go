package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

type AdminHandler struct {
	db *bolt.DB
}

func (h AdminHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hooks, err := h.listHooks()
	if err != nil {
		log.Print(err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	t := template.Must(template.ParseFiles("views/index.html"))
	if err := t.Execute(w, hooks); err != nil {
		log.Printf("error: %s", err)
	}
}

func (h AdminHandler) NewHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t := template.Must(template.ParseFiles("views/newhook.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Printf("error: %s", err)
	}
}

func (h AdminHandler) CreateHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")

	err := h.createHook(name)
	if err != nil {
		log.Printf("error creating hook: %s", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type Hook struct {
	Name string
}

func (h AdminHandler) listHooks() (hooks []Hook, err error) {
	err = h.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(BucketHooks).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			hooks = append(hooks, Hook{Name: string(k)})
		}
		return nil
	})
	return hooks, err
}

func (h AdminHandler) createHook(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("hook name is required")
	}

	return h.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketHooks)
		if b.Get([]byte(name)) != nil {
			return errors.New("a hook with that name already exists")
		}
		b.Put([]byte(name), nil)
		return nil
	})
}
