package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9000", "HTTP listen address")
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	log.Printf("Listening on %s", *listenAddr)
	log.Print(http.ListenAndServe(*listenAddr, mux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[r] %s %s", r.Method, r.RequestURI)
	fmt.Fprintf(w, "OK\n")
}
