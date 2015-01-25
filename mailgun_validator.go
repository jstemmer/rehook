package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("mailgun-validator", MailgunValidator{})
}

// MailgunValidator checks if the signature for an incoming request matches the
// calculated HMAC of the timestamp and token values. It also checks if the
// unique token hasn't been processed before to prevent replay attacks.
type MailgunValidator struct{}

// Name returns the name of this component.
func (MailgunValidator) Name() string { return "Mailgun validator" }

// Template returns the HTML template name of this component.
func (MailgunValidator) Template() string { return "mailgun-validator" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (MailgunValidator) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"apikey"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires a Mailgun API key to be
// present.
func (MailgunValidator) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	apikey, ok := params["apikey"]
	if !ok {
		return errors.New("apikey is required")
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-apikey", h.ID)), []byte(apikey)); err != nil {
		return err
	}
	_, err := b.CreateBucketIfNotExists([]byte("tokens"))
	return err
}

// Process verifies the signature and uniqueness of the random roken.
func (MailgunValidator) Process(h Hook, r Request, b *bolt.Bucket) error {
	// Check HMAC
	apikey := b.Get([]byte(fmt.Sprintf("%s-apikey", h.ID)))
	if apikey == nil {
		return errors.New("mailgun validator not initialized")
	}

	if r.Headers["Content-Type"] != "application/x-www-form-urlencoded" {
		return fmt.Errorf("unexpected Content-Type: %q", r.Headers["Content-Type"])
	}

	form, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return fmt.Errorf("error parsing request body: %s", err)
	}

	timestamp := form.Get("timestamp")
	signature := []byte(form.Get("signature"))
	token := form.Get("token")

	mac := hmac.New(sha256.New, apikey)
	mac.Write([]byte(timestamp + token))
	expected := []byte(hex.EncodeToString(mac.Sum(nil)))
	if !hmac.Equal(signature, expected) {
		return errors.New("invalid signature")
	}

	// Check uniqueness
	tokens := b.Bucket([]byte("tokens"))
	if p := tokens.Get([]byte(token)); p != nil {
		return errors.New("duplicate request token received")
	}
	return tokens.Put([]byte(token), []byte{})
}
