package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("execute-action", ExecuteAction{})
}

// ExecuteAction is a component that executes a command.
type ExecuteAction struct {
}

// Name returns the name of this component.
func (ExecuteAction) Name() string { return "Execute Command" }

// Template returns the HTML template name of this component.
func (ExecuteAction) Template() string { return "execute-action" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (ExecuteAction) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"command"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires a command to be present.
func (ExecuteAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	command, ok := params["command"]
	if !ok {
		return errors.New("command is required")
	}

	return b.Put([]byte(fmt.Sprintf("%s-command", h.ID)), []byte(command))
}

// Process executes command and logs the output and errors
func (ExecuteAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	command := b.Get([]byte(fmt.Sprintf("%s-command", h.ID)))
	if command == nil {
		return errors.New("forward request action not initialized")
	}

	out, err := exec.Command("sh", "-c", string(command)).CombinedOutput()
	log.Printf("[command-action][%s] executing command: %s\n --- STARTOUTPUT --- \n %s \n --- ENDOUTPUT ---", h.ID, command, out)
	if err != nil {
		log.Printf("[command-action][%s] error: %s", h.ID, err)
		return err
	}

	return nil
}
