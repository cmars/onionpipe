// Package secrets manages secret keys used in onion clients and servers.
package secrets

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ed25519"
)

// Secrets represents the format for storing oniongrok secret keys.
type Secrets struct {
	Version     string            `json:"version"`
	ServiceKeys map[string][]byte `json:"serviceKeys"`

	path    string
	changed bool
}

// ReadFile reads secrets from the given path.
func ReadFile(path string) (*Secrets, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), 0700)
		if err != nil {
			return nil, err
		}
		return &Secrets{
			Version: "1",
			path:    path,
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()
	sec, err := read(f)
	if err != nil {
		return nil, err
	}
	sec.path = path
	return sec, nil
}

func read(r io.Reader) (*Secrets, error) {
	var secrets Secrets
	err := json.NewDecoder(r).Decode(&secrets)
	if err != nil {
		return nil, err
	}
	return &secrets, nil
}

// WriteFile writes the secrets to the path from where they were read from, if
// they have changed.
func (s *Secrets) WriteFile() error {
	if !s.changed {
		return nil
	}
	if s.path == "" {
		return fmt.Errorf("don't know where to write")
	}
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.write(f)
}

func (s *Secrets) write(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

// EnsureServiceKey returns the service private key for the given alias name,
// generating a new one if it did not exist.
func (s *Secrets) EnsureServiceKey(name string) ([]byte, error) {
	if s.ServiceKeys == nil {
		s.ServiceKeys = map[string][]byte{}
	} else if key, ok := s.ServiceKeys[name]; ok {
		return key, nil
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	s.ServiceKeys[name] = []byte(priv)
	s.changed = true
	return []byte(priv), nil
}
