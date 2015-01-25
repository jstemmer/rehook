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

// WriteFileAction writes any incoming hook to a logfile in the logs/ dir.
type WriteFileAction struct {
}

// Name returns the name of this component.
func (WriteFileAction) Name() string { return "Write to file" }

// Template returns the HTML template name of this component.
func (WriteFileAction) Template() string { return "" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (WriteFileAction) Params(h Hook, b *bolt.Bucket) map[string]string {
	return nil
}

// Init initializes this component.
func (WriteFileAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	return nil
}

// Process writes a new file to the logs directory containing the headers and
// body of request r.
func (WriteFileAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return err
	}

	now := time.Now()
	if err := os.Mkdir("log", os.ModeDir); err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not create log directory: %s", err)
	}

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
