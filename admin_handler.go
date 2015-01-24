package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type AdminHandler struct {
	hooks *HookStore
}

func (h AdminHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hooks, err := h.hooks.List()
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
	hook := Hook{ID: r.FormValue("id")}
	if err := h.hooks.Create(hook); err != nil {
		// TODO: show flash message instead
		log.Printf("error creating hook: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}

func (h AdminHandler) EditHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	render("hooks/edit", w, hook)
}

func (h AdminHandler) DeleteHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := h.hooks.Delete(p.ByName("id")); err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
