package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type entry struct {
	userID    int64
	expiresAt time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]entry
	ttl      time.Duration
}

func NewStore(ttl time.Duration) *Store {
	s := &Store{
		sessions: make(map[string]entry),
		ttl:      ttl,
	}
	go s.cleanupLoop()
	return s
}

func (s *Store) Create(userID int64) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.sessions[id] = entry{userID: userID, expiresAt: time.Now().Add(s.ttl)}
	s.mu.Unlock()

	return id, nil
}

func (s *Store) cleanupLoop() {
	for range time.NewTicker(10 * time.Minute).C {
		now := time.Now()
		s.mu.Lock()

		for id, e := range s.sessions {
			if now.After(e.expiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func generateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
