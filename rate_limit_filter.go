package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("rate-limit-filter", RateLimitFilter{})
}

// RateLimitFilter limits the number of requests within a certain time interval.
type RateLimitFilter struct{}

// Name returns the name of this component.
func (RateLimitFilter) Name() string { return "Rate limiter" }

// Template returns the HTML template name of this component.
func (RateLimitFilter) Template() string { return "rate-limit-filter" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (RateLimitFilter) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"amount", "interval"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires an amount and an interval in
// seconds to be present.
func (RateLimitFilter) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	amount, ok := params["amount"]
	if !ok {
		return errors.New("amount is required")
	}

	if i, err := strconv.Atoi(amount); err != nil || i <= 0 {
		return fmt.Errorf("amount must be a positive number > 0: %s", err)
	}

	interval, ok := params["interval"]
	if !ok {
		return errors.New("interval is required")
	}

	if i, err := strconv.Atoi(interval); err != nil || i <= 0 {
		return fmt.Errorf("interval must be a positive number: %s", err)
	}

	if err := b.Put([]byte(fmt.Sprintf("%s-amount", h.ID)), []byte(amount)); err != nil {
		return err
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-interval", h.ID)), []byte(interval)); err != nil {
		return err
	}
	_, err := b.CreateBucketIfNotExists([]byte("requests"))
	return err
}

// Process makes sure incoming requests do not exceed the configured rate
// limit.
func (RateLimitFilter) Process(h Hook, r Request, b *bolt.Bucket) error {
	amount, _ := strconv.Atoi(string(b.Get([]byte(fmt.Sprintf("%s-amount", h.ID)))))
	interval, _ := strconv.Atoi(string(b.Get([]byte(fmt.Sprintf("%s-interval", h.ID)))))
	if amount <= 0 || interval <= 0 {
		return errors.New("rate limit filter not initialized")
	}

	b = b.Bucket([]byte("requests"))

	// store current timestamp
	now := time.Now()
	k := []byte(fmt.Sprintf("%d", now.UnixNano()))
	if err := b.Put(k, nil); err != nil {
		return err
	}

	// count requests
	c := b.Cursor()
	from := []byte(fmt.Sprintf("%d", now.Add(time.Duration(-interval)*time.Second).UnixNano()))

	var count int
	for k, _ := c.Seek(from); k != nil; k, _ = c.Next() {
		count++
	}

	if count > amount {
		return fmt.Errorf("rate limit exceeded (limit=%d count=%d)", amount, count)
	}

	// cleanup old entries
	for k, _ := c.First(); k != nil && bytes.Compare(k, from) <= 0; k, _ = c.Next() {
		if err := b.Delete(k); err != nil {
			return err
		}
	}
	return nil
}
