package main

import (
	"log"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("log-action", LogAction{})
}

type LogAction struct{}

func (LogAction) Name() string { return "Log" }

func (LogAction) Template() string { return "" }

func (LogAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	return nil
}

func (LogAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	log.Printf("[received] %s /h/%s", r.Method, h.ID)
	return nil
}
