package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("forward-request-action", ForwardRequestAction{})
}

// ForwardRequestAction is a component that forwards an incoming request to a
// user defined URL.
type ForwardRequestAction struct {
}

// Name returns the name of this component.
func (ForwardRequestAction) Name() string { return "Forward request" }

// Template returns the HTML template name of this component.
func (ForwardRequestAction) Template() string { return "request-forward-action" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (ForwardRequestAction) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"url"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires a valid url parameter to be
// present.
func (ForwardRequestAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	uri, ok := params["url"]
	if !ok {
		return errors.New("url is required")
	}

	if _, err := url.Parse(uri); err != nil {
		return fmt.Errorf("url is not valid: %s", err)
	}

	return b.Put([]byte(fmt.Sprintf("%s-url", h.ID)), []byte(uri))
}

// Process forwards the incoming request to the configured URL.
func (ForwardRequestAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	uri := b.Get([]byte(fmt.Sprintf("%s-url", h.ID)))
	if uri == nil {
		return errors.New("forward request action not initialized")
	}

	req, err := http.NewRequest(r.Method, string(uri), bytes.NewReader(r.Body))
	if err != nil {
		return fmt.Errorf("could not create new request: %s", err)
	}

	for k, v := range r.Headers {
		// special handling for some headers
		switch k {
		case "Connection":
			// skip
			continue
		case "User-Agent":
			// rename header, we will set our own user agent
			k = "X-Forwarded-User-Agent"
		}
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request forward error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("request forward unexpected status code received: %d", resp.StatusCode)
	}
	return nil
}
