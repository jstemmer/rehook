package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("write-file-action", WriteFileAction{})
}

// WriteFile writes any incoming hook to a logfile in dir.
type WriteFileAction struct {
}

func (WriteFileAction) Name() string { return "Write to file" }

func (WriteFileAction) Template() string { return "" }

func (WriteFileAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	return nil
}

// Process processes the incoming request r for hook h.
func (WriteFileAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return err
	}

	now := time.Now()
	filename := fmt.Sprintf("log/hook_%s_%s_%s.log", h.ID, now.Format("2006-01-02_15-04-05"), hex.EncodeToString(buf))

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Hook %q received at %s\n", h.ID, time.Now())
	fmt.Fprintf(f, "%s /h/%s\n\n", r.Method, h.ID)
	fmt.Fprintf(f, "Headers:\n")
	for k, v := range r.Headers {
		fmt.Fprintf(f, "%s = %s\n", k, v)
	}
	fmt.Fprintf(f, "\nBody:\n%s", r.Body)
	return err
}
