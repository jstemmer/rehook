package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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
		log.Print(err)
		http.NotFound(w, r)
		return
	}

	data := struct {
		Hook       *Hook
		Components map[string]Component
	}{hook, components}

	render("hooks/edit", w, data)
}

// UpdateHook handles POST requests from the edit page.
func (h AdminHandler) UpdateHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.FormValue("action") == "delete" {
		if err := h.hooks.Delete(p.ByName("id")); err != nil {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func (h AdminHandler) AddComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	id := r.URL.Query().Get("c")
	c, ok := components[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	tpl := c.Template()
	if tpl == "" {
		h.CreateComponent(w, r, p)
		return
	}
	render(fmt.Sprintf("components/%s", tpl), w, hook)
}

func (h AdminHandler) CreateComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	params := make(map[string]string)
	for k := range r.Form {
		if strings.HasPrefix(k, "param-") {
			params[k[6:]] = r.FormValue(k)
		}
	}

	if err := h.hooks.AddComponent(*hook, r.FormValue("c"), params); err != nil {
		// TODO: show flash message
		log.Printf("could not create component: %s", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}

func (h AdminHandler) EditComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// TODO: implement this
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h AdminHandler) UpdateComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	id := r.FormValue("c")
	action := r.FormValue("action")

	switch action {
	case "delete":
		if err := h.hooks.DeleteComponent(*hook, id); err != nil {
			log.Printf("error deleting component: %s", err)
		}
	case "move-up":
	case "move-down":
	default:
		// POST from edit page
	}

	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}
