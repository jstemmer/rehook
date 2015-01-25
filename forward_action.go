package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"net/http"
	"net/url"
)

func init() {
	RegisterComponent("forward-request-action", ForwardRequestAction{})
}

type ForwardRequestAction struct {
}

func (ForwardRequestAction) Name() string { return "Forward request" }

func (ForwardRequestAction) Template() string { return "request-forward-action" }

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

	req.Header.Set("User-Agent", "rehook/0.0.1 (https://github.com/gophergala/rehook)")

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
