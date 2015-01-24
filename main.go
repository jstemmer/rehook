package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9000", "HTTP listen address")
	adminAddr  = flag.String("admin", ":9001", "HTTP listen address for admin interface")
)

func main() {
	// webhooks
	mux := http.NewServeMux()
	mux.Handle("/h/", &HookHandler{})
	mux.HandleFunc("/", rootHandler)

	go func() {
		log.Printf("Listening on %s", *listenAddr)
		log.Print(http.ListenAndServe(*listenAddr, mux))
	}()

	// admin interface
	amux := http.NewServeMux()
	amux.Handle("/", &AdminHandler{})

	log.Printf("Admin interface on %s", *adminAddr)
	log.Print(http.ListenAndServe(*adminAddr, amux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.RequestURI != "/" {
		http.NotFound(w, r)
		return
	}
	log.Printf("[r] %s %s", r.Method, r.RequestURI)
	fmt.Fprintf(w, "OK\n")
}
