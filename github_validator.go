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

type GithubValidator struct{}

func (GithubValidator) Name() string { return "Github validator" }

func (GithubValidator) Template() string { return "github-validator" }

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
