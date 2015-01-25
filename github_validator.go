package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("github-validator", GithubValidator{})
}

// GithubValidator checks if the signature for an incoming request matches the
// calculated HMAC of the request body. It also checks if the unique identifier
// hasn't been processed before to prevent replay attacks.
type GithubValidator struct{}

// Name returns the name of this component.
func (GithubValidator) Name() string { return "Github validator" }

// Template returns the HTML template name of this component.
func (GithubValidator) Template() string { return "github-validator" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (GithubValidator) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"secret"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires a secret to be present.
func (GithubValidator) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	secret, ok := params["secret"]
	if !ok {
		return errors.New("secret is required")
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-secret", h.ID)), []byte(secret)); err != nil {
		return err
	}
	_, err := b.CreateBucketIfNotExists([]byte("deliveries"))
	return err
}

// Process verifies the signature and uniqueness of the delivery identifier.
func (GithubValidator) Process(h Hook, r Request, b *bolt.Bucket) error {
	// Check HMAC
	secret := b.Get([]byte(fmt.Sprintf("%s-secret", h.ID)))
	if secret == nil {
		return errors.New("github validator not initialized")
	}
	signature := []byte(r.Headers["X-Hub-Signature"])

	mac := hmac.New(sha1.New, secret)
	mac.Write(bytes.TrimSpace(r.Body))
	expected := append([]byte("sha1="), hex.EncodeToString(mac.Sum(nil))...)
	if !hmac.Equal(signature, expected) {
		return errors.New("invalid signature")
	}

	// Check uniqueness
	id := []byte(r.Headers["X-Github-Delivery"])
	deliveries := b.Bucket([]byte("deliveries"))
	if did := deliveries.Get([]byte(id)); did != nil {
		return errors.New("duplicate delivery")
	}
	return deliveries.Put([]byte(id), []byte{})
}
