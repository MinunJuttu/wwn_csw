package adminauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"
)

const Lifetime = 12 * time.Hour

type Store struct {
	mu       sync.RWMutex
	sessions map[[32]byte]time.Time
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[[32]byte]time.Time),
	}
}

func (s *Store) Create() (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().Add(Lifetime)

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, expiry := range s.sessions {
		if time.Now().After(expiry) {
			delete(s.sessions, key)
		}
	}

	s.sessions[hash] = expiresAt

	return token, expiresAt, nil
}

func (s *Store) Valid(token string) bool {
	if token == "" {
		return false
	}

	hash := sha256.Sum256([]byte(token))

	s.mu.RLock()
	expiresAt, ok := s.sessions[hash]
	s.mu.RUnlock()

	if !ok {
		return false
	}

	if time.Now().After(expiresAt) {
		s.mu.Lock()
		delete(s.sessions, hash)
		s.mu.Unlock()
		return false
	}

	return true
}

func (s *Store) Delete(token string) {
	if token == "" {
		return
	}

	hash := sha256.Sum256([]byte(token))

	s.mu.Lock()
	delete(s.sessions, hash)
	s.mu.Unlock()
}
