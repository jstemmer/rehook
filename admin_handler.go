package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	data := struct {
		ID  string
		Err string
	}{r.URL.Query().Get("id"), r.URL.Query().Get("err")}
	render("hooks/new", w, data)
}

// CreateHook handles POST requests from the new hook form.
func (h AdminHandler) CreateHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.FormValue("id")
	hook := Hook{ID: id}
	if err := h.hooks.Create(hook); err != nil {
		log.Printf("error creating hook: %s", err)
		// TODO: maybe use sessions for flash messages etc
		http.Redirect(w, r, fmt.Sprintf("/hooks/new?id=%s&err=%s", url.QueryEscape(id), url.QueryEscape(err.Error())), http.StatusSeeOther)
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

// AddComponent renders the component configuration screen, if it has one, or
// directly redirects to h.CreateComponent.
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
	data := struct {
		ID     string
		Hook   *Hook
		Params map[string]string
	}{"", hook, map[string]string{"interval": ""}}
	render(fmt.Sprintf("components/%s", tpl), w, data)
}

// CreateComponent adds a new instance of the selected component to the current
// hook.
func (h AdminHandler) CreateComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	params := filterParams(r)

	if err := h.hooks.AddComponent(*hook, r.FormValue("c"), params); err != nil {
		// TODO: show flash message
		log.Printf("could not create component: %s", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}

func filterParams(r *http.Request) map[string]string {
	r.ParseForm()
	params := make(map[string]string)
	for k := range r.Form {
		if strings.HasPrefix(k, "param-") {
			params[k[6:]] = r.FormValue(k)
		}
	}
	return params
}

// EditComponent renders the component configuration page.
func (h AdminHandler) EditComponent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var cid string
	var c Component
	id := p.ByName("c")
	for _, hc := range hook.Components {
		if hc.ID == id {
			cid = hc.Name
			c = components[hc.Name]
			break
		}
	}

	if c == nil || c.Template() == "" {
		http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
		return
	}

	params, err := h.hooks.ComponentParams(*hook, cid)
	if err != nil {
		log.Printf("error: %s", err)
		http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
		return
	}

	data := struct {
		ID     string
		Hook   *Hook
		Params map[string]string
	}{id, hook, params}
	render(fmt.Sprintf("components/%s", c.Template()), w, data)
}

// UpdateComponent handles updates to a component instance. This includes
// moving the processing order or deleting the component.
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
		// TODO: implement this
	case "move-down":
		// TODO: implement this
	default:
		// TODO: POST from edit page; update params
		params := filterParams(r)
		if err := h.hooks.UpdateComponent(*hook, id, params); err != nil {
			log.Printf("error updating component: %s", err)
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/hooks/edit/%s", hook.ID), http.StatusSeeOther)
}
