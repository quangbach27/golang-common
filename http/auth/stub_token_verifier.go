package auth

import (
	"context"
	"errors"
	"sync"
)

var ErrInvalidToken = errors.New("invalid token")

// StubTokenVerifier is an in-memory TokenVerifier for tests and local development.
// Do not use it in production.
type StubTokenVerifier struct {
	mu       sync.RWMutex
	sessions map[string]*Session

	// Err, if set, is returned for every Verify call (simulates verifier outage).
	Err error
}

var _ TokenVerifier = (*StubTokenVerifier)(nil)

func NewStubTokenVerifier() *StubTokenVerifier {
	return &StubTokenVerifier{sessions: make(map[string]*Session)}
}

// Add registers a token that resolves to the given session.
func (v *StubTokenVerifier) Add(token string, s *Session) *StubTokenVerifier {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.sessions[token] = s
	return v
}

func (v *StubTokenVerifier) Verify(_ context.Context, token string) (*Session, error) {
	if v.Err != nil {
		return nil, v.Err
	}

	v.mu.RLock()
	s, ok := v.sessions[token]
	v.mu.RUnlock()
	if !ok {
		return nil, ErrInvalidToken
	}

	return s, nil
}
