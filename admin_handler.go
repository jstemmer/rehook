package main

import (
	"fmt"
	"net/http"
)

type AdminHandler struct {
}

func (h AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html><h2>rehook admin</h2></html>")
}
