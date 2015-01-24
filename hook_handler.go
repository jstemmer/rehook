package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

// HookHandler is the webhook HTTP handler.
type HookHandler struct {
	hooks *HookStore
	db    *bolt.DB
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

	req, err := loadRequest(r)
	if err != nil {
		log.Printf("error reading request: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	go h.processRequest(hook, req)
	w.WriteHeader(http.StatusOK)
}

func (h *HookHandler) processRequest(hook *Hook, r Request) {
	for i, c := range hook.Components {
		// TODO: remove debug logging
		log.Printf("%d: processing %s", i+1, c.Name)

		cmp, ok := components[c.Name]
		if !ok {
			log.Printf("skipping unknown component: %s", c.Name)
			continue
		}

		tx, err := h.db.Begin(true)
		if err != nil {
			log.Printf("error starting db tx: %s", err)
			break
		}

		b := tx.Bucket(BucketComponents).Bucket([]byte(c.Name))
		if err := cmp.Process(*hook, r, b); err != nil {
			tx.Rollback()
			log.Printf("processing stopped: %s", err)
			break
		}
		if err := tx.Commit(); err != nil {
			log.Printf("error committing db tx: %s", err)
			break
		}
	}

	if err := h.hooks.Inc(hook.ID); err != nil {
		log.Printf("error incrementing count for %s: %s", hook.ID, err)
	}
}
