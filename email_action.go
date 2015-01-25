package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/boltdb/bolt"
)

func init() {
	RegisterComponent("email-action", EmailAction{})
}

// EmailAction is a component that sends an email using the Mailgun API.
type EmailAction struct {
}

// Name returns the name of this component.
func (EmailAction) Name() string { return "Send email (using Mailgun)" }

// Template returns the HTML template name of this component.
func (EmailAction) Template() string { return "email-action" }

// Params returns the currently stored configuration parameters for hook h
// from bucket b.
func (EmailAction) Params(h Hook, b *bolt.Bucket) map[string]string {
	m := make(map[string]string)
	for _, k := range []string{"token", "domain", "address", "subject", "template"} {
		m[k] = string(b.Get([]byte(fmt.Sprintf("%s-%s", h.ID, k))))
	}
	return m
}

// Init initializes this component. It requires a Mailgun token, a valid
// address and template parameter to be present.
func (EmailAction) Init(h Hook, params map[string]string, b *bolt.Bucket) error {
	// TODO: refactor the following
	token, ok := params["token"]
	if !ok {
		return errors.New("token is required")
	}

	domain, ok := params["domain"]
	if !ok {
		return errors.New("domain is required")
	}

	address, ok := params["address"]
	if !ok {
		return errors.New("address is required")
	}

	subject, ok := params["subject"]
	if !ok {
		return errors.New("subject is required")
	}

	tpl, ok := params["template"]
	if !ok {
		return errors.New("template is required")
	}

	if _, err := template.New("email").Parse(string(tpl)); err != nil {
		return fmt.Errorf("invalid template: %s", err)
	}

	if err := b.Put([]byte(fmt.Sprintf("%s-token", h.ID)), []byte(token)); err != nil {
		return err
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-domain", h.ID)), []byte(domain)); err != nil {
		return err
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-address", h.ID)), []byte(address)); err != nil {
		return err
	}
	if err := b.Put([]byte(fmt.Sprintf("%s-subject", h.ID)), []byte(subject)); err != nil {
		return err
	}
	return b.Put([]byte(fmt.Sprintf("%s-template", h.ID)), []byte(tpl))
}

// Process sends an email to the configured address using the template as email
// body.
func (EmailAction) Process(h Hook, r Request, b *bolt.Bucket) error {
	token := b.Get([]byte(fmt.Sprintf("%s-token", h.ID)))
	domain := b.Get([]byte(fmt.Sprintf("%s-domain", h.ID)))
	address := b.Get([]byte(fmt.Sprintf("%s-address", h.ID)))
	subject := b.Get([]byte(fmt.Sprintf("%s-subject", h.ID)))
	tpl := b.Get([]byte(fmt.Sprintf("%s-template", h.ID)))
	if token == nil || domain == nil || address == nil || subject == nil || tpl == nil {
		return errors.New("email action not initialized")
	}

	t, err := template.New("email").Parse(string(tpl))
	if err != nil {
		return fmt.Errorf("could not parse template: %s", err)
	}

	data := struct {
		Hook    Hook
		Request Request
	}{h, r}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return fmt.Errorf("could not execute template: %s", err)
	}
	return sendMail(string(token), string(domain), string(address), string(subject), buf.String())
}

func sendMail(token, domain, address, subject, text string) error {
	form := url.Values{}
	form.Set("from", "mail@"+domain)
	form.Set("to", address)
	form.Set("subject", subject)
	form.Set("text", text)

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.mailgun.net/v2/%s/messages", domain), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", token)

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			log.Printf("Mailgun API error response: %s", body)
		}
		return fmt.Errorf("send mail unexpected status code received: %d", resp.StatusCode)
	}
	return nil
}
