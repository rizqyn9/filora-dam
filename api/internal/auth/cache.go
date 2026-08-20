package auth

import (
	"sync"
	"time"
)

// CachedSession holds session data in the in-process LRU cache.
type CachedSession struct {
	SessionID int64
	UserID    int64
	ExpiresAt time.Time
}

// TokenCache is an in-process cache for session tokens.
// For 3-5 users, all active tokens fit in memory permanently.
type TokenCache struct {
	mu    sync.RWMutex
	items map[string]*CachedSession // key = token_hash
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		items: make(map[string]*CachedSession),
	}
}

// Get returns a cached session if it exists and is not expired.
func (c *TokenCache) Get(tokenHash string) (*CachedSession, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok := c.items[tokenHash]
	if !ok {
		return nil, false
	}
	if time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

// Set stores a session in the cache.
func (c *TokenCache) Set(tokenHash string, session *CachedSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[tokenHash] = session
}

// Delete removes a session from the cache (on revocation).
func (c *TokenCache) Delete(tokenHash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, tokenHash)
}

// DeleteByUserID removes all sessions for a user (on logout-all).
func (c *TokenCache) DeleteByUserID(userID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.UserID == userID {
			delete(c.items, k)
		}
	}
}
