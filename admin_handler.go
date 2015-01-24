package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	render("hooks/index", w, hooks)
}

func (h AdminHandler) NewHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	render("hooks/new", w, nil)
}

func (h AdminHandler) CreateHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")
	if err := h.createHook(name); err != nil {
		// TODO: show flash message instead
		log.Printf("error creating hook: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/hooks/%s", name), http.StatusSeeOther)
}

func (h AdminHandler) EditHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")
	hook, err := h.findHook(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	render("hooks/edit", w, hook)
}

func (h AdminHandler) DeleteHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")
	if err := h.deleteHook(name); err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type Hook struct {
	ID string
}

func (h AdminHandler) listHooks() (hooks []Hook, err error) {
	err = h.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(BucketHooks).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			hooks = append(hooks, Hook{ID: string(k)})
		}
		return nil
	})
	return hooks, err
}

func (h AdminHandler) findHook(name string) (hook *Hook, err error) {
	err = h.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(BucketHooks).Get([]byte(name))
		if value == nil {
			return errors.New("hook does not exist")
		}
		hook = &Hook{name}
		return nil
	})
	return hook, err
}

func (h AdminHandler) createHook(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("hook name is required")
	}

	if match, err := regexp.MatchString("^[a-z0-9-]+$", name); err != nil || !match {
		if err != nil {
			log.Printf("create hook regexp error: %s", err)
		}
		return errors.New("name contains invalid characters")
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

func (h AdminHandler) deleteHook(name string) error {
	return h.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BucketHooks).Delete([]byte(name))
	})
}
