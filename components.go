package main

import (
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
)

var (
	components = make(map[string]Component)
)

// RegisterComponent registers component c with name as the unique identifier
// for this component.
func RegisterComponent(name string, c Component) {
	if c == nil {
		panic("cannot register nil component")
	}
	components[name] = c
}

// HookComponent is a component that belongs to an existing hook.
type HookComponent struct {
	ID   string
	Name string
}

// Request represents an incoming request that may be processed by components.
type Request struct {
	Method  string
	Headers map[string]string
	Body    []byte
}

func loadRequest(req *http.Request) (r Request, err error) {
	r.Method = req.Method
	r.Headers = make(map[string]string)
	for k := range req.Header {
		r.Headers[k] = req.Header.Get(k)
	}
	r.Body, err = ioutil.ReadAll(req.Body)
	return r, err
}

// Component is the interface all components must implement. Hooks may have
// many components, each processing an incoming request.
type Component interface {
	// Name returns the human-friendly name of the component.
	Name() string

	// Template returns the name of the template to render when configuring the
	// component. An empty string indicates no configuration is needed.
	Template() string

	// Params returns the currently stored configuration parameters for hook h
	// from bucket b.
	Params(h Hook, b *bolt.Bucket) map[string]string

	// Init is called when this component is added to a hook. Params contains a
	// map of user configured settings. If this component could not be
	// initialized, a descriptive error should be returned. The bucket b can be
	// used to store data. Note that the bucket is shared between all instances
	// of this component.
	Init(h Hook, params map[string]string, b *bolt.Bucket) error

	// Process is called whenever an incoming request r passes through this
	// component for an existing hook h. Bucket b is provided to fetch or store
	// data. If this request cannot be processed, a descriptive error should be
	// returned.
	Process(h Hook, r Request, b *bolt.Bucket) error
}
