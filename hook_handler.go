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

	"github.com/julienschmidt/httprouter"
)

// HookHandler is the webhook HTTP handler.
type HookHandler struct {
	hooks *HookStore
}

// ReceiveHook handles incoming webhook HTTP requests.
func (h *HookHandler) ReceiveHook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	hook, err := h.hooks.Find(p.ByName("id"))
	if err != nil {
		log.Printf("no hook configured for %q", id)
		http.NotFound(w, r)
		return
	}

	wf := &WriteFile{"log"}
	if err := wf.Process(hook, r); err != nil {
		log.Printf("error processing hook %s: %s", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[received] %s %s", r.Method, r.RequestURI)
	w.WriteHeader(http.StatusOK)
}

// WriteFile writes any incoming hook to a logfile in dir.
type WriteFile struct {
	dir string
}

// Process processes the incoming request r for hook h.
func (w WriteFile) Process(h *Hook, r *http.Request) error {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return err
	}

	now := time.Now()
	filename := fmt.Sprintf("%s/hook_%s_%s_%s.log", w.dir, h.ID, now.Format("2006-01-02_15-04-05"), hex.EncodeToString(buf))

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Hook %q received at %s\n", h.ID, time.Now())
	fmt.Fprintf(f, "From %v\n\n", r.RemoteAddr)
	fmt.Fprintf(f, "%s %s\n\n", r.Method, r.RequestURI)
	fmt.Fprintf(f, "Headers:\n")
	for k, v := range r.Header {
		fmt.Fprintf(f, "%s = %v\n", k, v)
	}
	fmt.Fprintf(f, "\nBody:\n")
	_, err = io.Copy(f, r.Body)
	return err
}
