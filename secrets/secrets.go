// Package secrets manages secret keys used in onion clients and servers.
package secrets

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

// Secrets represents the format for storing onionpipe secret keys.
type Secrets struct {
	Version     string                   `json:"version"`
	ServiceKeys map[string][]byte        `json:"serviceKeys"`
	ClientKeys  map[string]ClientKeyPair `json:"clientKeys"`

	path    string
	changed bool
}

// ClientKeyPair represents an x25519 key pair used for client authorization.
type ClientKeyPair struct {
	Public  []byte `json:"public"`
	Private []byte `json:"private"`
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

// RemoveServiceKey removes the service private key for the given alias name.
func (s *Secrets) RemoveServiceKey(name string) error {
	if _, ok := s.ServiceKeys[name]; !ok {
		return fmt.Errorf("key %q not found", name)
	}
	delete(s.ServiceKeys, name)
	s.changed = true
	return nil
}

// ServicesPublic represent public key information about services.
type ServicesPublic map[string]ServicePublic

// ServicePublic represents public key information about a service.
type ServicePublic struct {
	Address string `json:"address"`
}

// ServicesPublic returns public key information about the service keys.
func (s *Secrets) ServicesPublic() ServicesPublic {
	services := ServicesPublic{}
	for name, serviceKey := range s.ServiceKeys {
		pubKey := ed25519.PrivateKey(serviceKey).Public().(ed25519.PublicKey)
		services[name] = ServicePublic{
			Address: strings.ToLower(
				base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(pubKey[:])),
		}
	}
	return services
}

// EnsureClientKey returns the client private key for the given alias name,
// generating a new one if it did not exist.
func (s *Secrets) EnsureClientKey(name string) (ClientKeyPair, error) {
	if s.ClientKeys == nil {
		s.ClientKeys = map[string]ClientKeyPair{}
	} else if key, ok := s.ClientKeys[name]; ok {
		return key, nil
	}
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return ClientKeyPair{}, err
	}
	keyPair := ClientKeyPair{
		Public:  pub[:],
		Private: priv[:],
	}
	s.ClientKeys[name] = keyPair
	s.changed = true
	return keyPair, nil
}

// ResolveClientPrivateKey returns the x25519 client authorization private key
// for the given identity name, or base32-encoded private key representation.
func (s *Secrets) ResolveClientPrivateKey(nameOrKey string) ([]byte, error) {
	if key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(nameOrKey)); err == nil {
		if len(key) == 32 {
			return key, nil
		}
	}
	if keyPair, ok := s.ClientKeys[nameOrKey]; ok {
		return keyPair.Private, nil
	}
	return nil, fmt.Errorf("failed to resolve client key %q", nameOrKey)
}

// RemoveClientKey removes the client private key for the given alias name.
func (s *Secrets) RemoveClientKey(name string) error {
	if _, ok := s.ClientKeys[name]; !ok {
		return fmt.Errorf("key %q not found", name)
	}
	delete(s.ClientKeys, name)
	s.changed = true
	return nil
}

// ClientsPublic represent client public key identities that can be authorized
// to access onion services.
type ClientsPublic map[string]ClientPublic

// ClientPublic represents a client public key identity, that can be authorized
// to access onion services.
type ClientPublic struct {
	Identity string `json:"identity"`
}

// ClientsPublic returns public key information about the client identities
// held in this secret store. Clients' public keys would be shared with the
// operator of an authenticated onion service, for granting exclusive access to
// these key identities.
func (s *Secrets) ClientsPublic() ClientsPublic {
	clients := ClientsPublic{}
	for name, keyPair := range s.ClientKeys {
		clients[name] = ClientPublic{
			Identity: strings.ToLower(
				base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(keyPair.Public)),
		}
	}
	return clients
}
