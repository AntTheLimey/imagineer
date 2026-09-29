/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package conversation

import (
	"sync"
	"time"

	"github.com/antonypegg/imagineer/internal/llm"
)

// CacheEntry holds in-memory conversation state for a
// single conversation, including the accumulated messages,
// system prompt, and tool definitions.
type CacheEntry struct {
	Messages     []llm.StreamingMessage
	SystemPrompt string
	ToolDefs     []llm.ToolDefinition
	LastAccessed time.Time
}

// SessionCache provides idle-evicting conversation state.
// Entries that have not been accessed within the configured
// TTL are automatically removed by a background sweeper.
type SessionCache struct {
	mu      sync.Mutex
	entries map[int64]*CacheEntry
	ttl     time.Duration
	stopCh  chan struct{}
}

// NewSessionCache creates a new cache and starts a
// background goroutine that sweeps expired entries at
// an interval of ttl/2. Call Stop to shut down the
// sweeper when the cache is no longer needed.
func NewSessionCache(ttl time.Duration) *SessionCache {
	c := &SessionCache{
		entries: make(map[int64]*CacheEntry),
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}
	go c.sweepLoop()
	return c
}

// Get returns the cache entry for the given conversation
// ID, updating LastAccessed to extend the TTL. Returns
// nil, false if the entry does not exist or has expired.
//
// The caller must not modify the returned entry
// concurrently with other goroutines accessing the same
// conversation. Use Set to store updated state.
func (c *SessionCache) Get(convID int64) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[convID]
	if !ok {
		return nil, false
	}
	if time.Since(entry.LastAccessed) > c.ttl {
		delete(c.entries, convID)
		return nil, false
	}
	entry.LastAccessed = time.Now()
	return entry, true
}

// Set stores a cache entry for the given conversation ID
// and sets its LastAccessed timestamp to the current time.
func (c *SessionCache) Set(convID int64, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry.LastAccessed = time.Now()
	c.entries[convID] = entry
}

// Invalidate removes the cache entry for the given
// conversation ID.
func (c *SessionCache) Invalidate(convID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, convID)
}

// Stop shuts down the background sweeper goroutine.
// It is safe to call Stop multiple times, but only the
// first call has any effect.
func (c *SessionCache) Stop() {
	select {
	case <-c.stopCh:
		// Already stopped.
	default:
		close(c.stopCh)
	}
}

// sweepLoop runs in a background goroutine, periodically
// removing entries that have exceeded the TTL.
func (c *SessionCache) sweepLoop() {
	interval := c.ttl / 2
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.sweep()
		}
	}
}

// sweep removes all entries whose LastAccessed time
// exceeds the configured TTL.
func (c *SessionCache) sweep() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.entries {
		if now.Sub(entry.LastAccessed) > c.ttl {
			delete(c.entries, id)
		}
	}
}
