package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

var (
	listenAddr = flag.String("http", ":9000", "HTTP listen address")
	adminAddr  = flag.String("admin", ":9001", "HTTP listen address for admin interface")
)

func main() {
	// initialize database
	db, err := bolt.Open("data.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Could not open database: %s", err)
	}
	defer db.Close()

	// webhooks
	mux := http.NewServeMux()
	mux.Handle("/h/", &HookHandler{})
	mux.Handle("/", http.NotFoundHandler())

	go func() {
		log.Printf("Listening on %s", *listenAddr)
		log.Print(http.ListenAndServe(*listenAddr, mux))
	}()

	// admin interface
	ah := &AdminHandler{}
	arouter := httprouter.New()
	arouter.GET("/", ah.Index)

	log.Printf("Admin interface on %s", *adminAddr)
	log.Print(http.ListenAndServe(*adminAddr, arouter))
}
