package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type AdminHandler struct {
}

func (h AdminHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t := template.Must(template.ParseFiles("views/index.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Printf("error: %s", err)
	}
}
