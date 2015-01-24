package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

type HookHandler struct {
}

func (h *HookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %s", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	log.Printf("-- hook received --")
	log.Printf("%s %s", r.Method, r.RequestURI)
	for k, v := range r.Header {
		log.Printf("%s = %s", k, v)
	}
	log.Println(string(body))
}
