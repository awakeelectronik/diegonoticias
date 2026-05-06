package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type Session struct {
	Username  string
	CSRFToken string
	ExpiresAt time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

func (m *SessionManager) Create(username string) (token string, csrf string, expiresAt time.Time, err error) {
	token, err = randomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}
	csrf, err = randomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}
	expiresAt = time.Now().Add(m.ttl)
	m.mu.Lock()
	m.sessions[token] = Session{
		Username:  username,
		CSRFToken: csrf,
		ExpiresAt: expiresAt,
	}
	m.mu.Unlock()
	return token, csrf, expiresAt, nil
}

func (m *SessionManager) Get(token string) (Session, bool) {
	m.mu.RLock()
	s, ok := m.sessions[token]
	m.mu.RUnlock()
	if !ok {
		return Session{}, false
	}
	if time.Now().After(s.ExpiresAt) {
		m.Delete(token)
		return Session{}, false
	}
	return s, true
}

func (m *SessionManager) Delete(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

