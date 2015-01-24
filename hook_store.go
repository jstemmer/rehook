package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// HookStore is the database that stores hook configuration and data.
type HookStore struct {
	db *bolt.DB
}

// Hook is the configuration for a single hook.
type Hook struct {
	ID string // unique hook identifier
}

// List returns a list of all hooks.
func (s *HookStore) List() (hooks []Hook, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(BucketHooks).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			hooks = append(hooks, Hook{ID: string(k)})
		}
		return nil
	})
	return hooks, err
}

// Find returns the hook with the given id if it exists, nil otherwise.
func (s *HookStore) Find(id string) (h *Hook, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(BucketHooks).Get([]byte(id))
		if value == nil {
			return errors.New("hook does not exist")
		}
		h = &Hook{id}
		return nil
	})
	return h, err
}

// Create creates hook h.
func (s *HookStore) Create(h Hook) error {
	if strings.TrimSpace(h.ID) == "" {
		return errors.New("hook id is required")
	}

	if match, err := regexp.MatchString("^[a-z0-9-]+$", h.ID); err != nil || !match {
		if err != nil {
			log.Printf("create hook regexp error: %s", err)
		}
		return errors.New("hook id contains invalid characters")
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketHooks)
		if b.Get([]byte(h.ID)) != nil {
			return errors.New("a hook with that id already exists")
		}
		b.Put([]byte(h.ID), nil)
		return nil
	})
}

// Delete deletes the hook with the given id if it exists.
func (s *HookStore) Delete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BucketHooks).Delete([]byte(id))
	})
}

// Inc increments the count for the hook with the given id.
func (s *HookStore) Inc(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketStats)
		// increment; group by date and hour
		err := increment([]byte(fmt.Sprintf("%s-%s", id, time.Now().Format("2006-01-02-15"))), b)
		if err != nil {
			return err
		}
		return increment([]byte(fmt.Sprintf("%s-total", id)), b)
	})
}

func increment(key []byte, b *bolt.Bucket) (err error) {
	var count int

	var v []byte
	if v = b.Get(key); v != nil {
		if err = gobDecode(v, &count); err != nil {
			return err
		}
	}
	count++
	if v, err = gobEncode(count); err != nil {
		return err
	}
	log.Printf("incremented %v to %d", string(key), count)
	return b.Put(key, v)
}

func gobEncode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}

func gobDecode(p []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(p)).Decode(v)
}
