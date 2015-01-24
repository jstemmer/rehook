package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AdminHandler handles requests for the admin web interface.
type AdminHandler struct {
	hooks *HookStore
}

// Index renders the main page that shows a list of hooks.
func (h AdminHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hooks, err := h.hooks.List()
	if err != nil {
		log.Print(err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	render("hooks/index", w, hooks)
}

// NewHook renders the new hook form.
func (h AdminHandler) NewHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	render("hooks/new", w, nil)
}

// CreateHook handles POST requests from the new hook form.
func (h AdminHandler) CreateHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hook := Hook{ID: r.FormValue("id")}
	if err := h.hooks.Create(hook); err != nil {
		// TODO: show flash message instead
		log.Printf("error creating hook: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}

// EditHook renders the edit hook form
func (h AdminHandler) EditHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	render("hooks/edit", w, hook)
}

// DeleteHook handles POST requests to delete a hook
func (h AdminHandler) DeleteHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := h.hooks.Delete(p.ByName("id")); err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
