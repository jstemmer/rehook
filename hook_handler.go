package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

type HookHandler struct {
	db *bolt.DB
}

func (h *HookHandler) ReceiveHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	// log to file for now
	f, err := os.Create(filename(id))
	if err != nil {
		log.Printf("error creating file: %s", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "Hook %q received at %s\n", id, time.Now())
	fmt.Fprintf(f, "From %v\n\n", r.RemoteAddr)
	fmt.Fprintf(f, "%s %s\n\n", r.Method, r.RequestURI)
	fmt.Fprintf(f, "Headers:\n")
	for k, v := range r.Header {
		fmt.Fprintf(f, "%s = %v\n", k, v)
	}
	fmt.Fprintf(f, "\nBody:\n")
	if _, err := io.Copy(f, r.Body); err != nil {
		log.Printf("error copying request body to file: %s", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Printf("[received] %s %s", r.Method, r.RequestURI)
}

func filename(id string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return fmt.Sprintf("log/hook_%s_%s_%s.log", id, time.Now().Format("2006-01-02_15-04-05"), hex.EncodeToString(buf))
}
