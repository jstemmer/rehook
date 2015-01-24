package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

// flags
var (
	listenAddr = flag.String("http", ":9000", "HTTP listen address")
	adminAddr  = flag.String("admin", ":9001", "HTTP listen address for admin interface")
)

// Database constants
var (
	BucketHooks = []byte("hooks")
	BucketStats = []byte("stats")
)

func main() {
	// initialize database
	db, err := bolt.Open("data.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Could not open database: %s", err)
	}
	defer db.Close()

	if err := db.Update(initBuckets); err != nil {
		log.Fatal(err)
	}

	// webhooks
	mux := http.NewServeMux()
	mux.Handle("/h/", &HookHandler{})
	mux.Handle("/", http.NotFoundHandler())

	go func() {
		log.Printf("Listening on %s", *listenAddr)
		log.Print(http.ListenAndServe(*listenAddr, mux))
	}()

	// admin interface
	ah := &AdminHandler{db}
	arouter := httprouter.New()
	arouter.GET("/", ah.Index)
	arouter.GET("/hooks/new", ah.NewHook)
	arouter.POST("/hooks", ah.CreateHook)

	log.Printf("Admin interface on %s", *adminAddr)
	log.Print(http.ListenAndServe(*adminAddr, arouter))
}

func initBuckets(t *bolt.Tx) error {
	for _, name := range [][]byte{BucketHooks, BucketStats} {
		if _, err := t.CreateBucketIfNotExists(name); err != nil {
			return err
		}
	}
	return nil
}
