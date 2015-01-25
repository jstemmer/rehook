package main

import (
	"log"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("log-action", LogAction{})
}

// LogAction is a logger component.
type LogAction struct{}

// Name returns the name of this component.
func (LogAction) Name() string { return "Log" }

// Template returns the HTML template name of this component.
func (LogAction) Template() string { return "" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (LogAction) Params(h Hook, b *bolt.Bucket) map[string]string {
	return nil
}

// Init initializes this component.
func (LogAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	return nil
}

// Process writes the current request method and hook id to stderr.
func (LogAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	log.Printf("[received] %s /h/%s", r.Method, h.ID)
	return nil
}
