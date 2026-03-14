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
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestSessionCacheGetSet(t *testing.T) {
	cache := NewSessionCache(30 * time.Minute)
	defer cache.Stop()

	entry := &CacheEntry{
		Messages: []llm.StreamingMessage{
			{Role: "user", Content: "Hello"},
		},
	}
	cache.Set(42, entry)

	got, ok := cache.Get(42)
	assert.True(t, ok)
	assert.Len(t, got.Messages, 1)
	assert.Equal(t, "user", got.Messages[0].Role)
	assert.Equal(t, "Hello", got.Messages[0].Content)
}

func TestSessionCacheEviction(t *testing.T) {
	cache := NewSessionCache(10 * time.Millisecond)
	defer cache.Stop()

	cache.Set(1, &CacheEntry{})
	time.Sleep(50 * time.Millisecond)

	_, ok := cache.Get(1)
	assert.False(t, ok)
}

func TestSessionCacheInvalidate(t *testing.T) {
	cache := NewSessionCache(30 * time.Minute)
	defer cache.Stop()

	cache.Set(1, &CacheEntry{})
	cache.Invalidate(1)

	_, ok := cache.Get(1)
	assert.False(t, ok)
}

func TestSessionCacheGetUpdatesLastAccessed(t *testing.T) {
	cache := NewSessionCache(100 * time.Millisecond)
	defer cache.Stop()

	cache.Set(1, &CacheEntry{
		Messages: []llm.StreamingMessage{
			{Role: "user", Content: "ping"},
		},
	})

	// Repeatedly access the entry before the TTL expires
	// to confirm that Get refreshes the deadline.
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		got, ok := cache.Get(1)
		assert.True(t, ok, "entry should survive on iteration %d", i)
		assert.Len(t, got.Messages, 1)
	}

	// Total elapsed time is ~150ms, well past the 100ms TTL,
	// but the entry should still be alive because each Get
	// refreshed LastAccessed.
	got, ok := cache.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "ping", got.Messages[0].Content)
}

func TestSessionCacheStop(t *testing.T) {
	// Record baseline goroutine count before creating the
	// cache, allowing a small buffer for runtime jitter.
	baseline := runtime.NumGoroutine()

	cache := NewSessionCache(time.Hour)
	cache.Set(1, &CacheEntry{})

	// The sweeper goroutine should be running.
	afterCreate := runtime.NumGoroutine()
	assert.Greater(t, afterCreate, baseline,
		"sweeper goroutine should be running")

	cache.Stop()

	// Give the goroutine time to exit.
	time.Sleep(20 * time.Millisecond)

	afterStop := runtime.NumGoroutine()
	assert.LessOrEqual(t, afterStop, baseline+1,
		"sweeper goroutine should have exited")

	// Calling Stop again should not panic.
	cache.Stop()
}

func TestSessionCacheConcurrency(t *testing.T) {
	cache := NewSessionCache(time.Second)
	defer cache.Stop()

	var wg sync.WaitGroup
	const workers = 20
	const iterations = 200

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			convID := int64(id % 10)
			for i := 0; i < iterations; i++ {
				cache.Set(convID, &CacheEntry{
					Messages: []llm.StreamingMessage{
						{Role: "user", Content: "msg"},
					},
				})
				cache.Get(convID)
				if i%3 == 0 {
					cache.Invalidate(convID)
				}
			}
		}(w)
	}
	wg.Wait()
}

func TestSessionCacheGetMiss(t *testing.T) {
	cache := NewSessionCache(30 * time.Minute)
	defer cache.Stop()

	got, ok := cache.Get(999)
	assert.False(t, ok)
	assert.Nil(t, got)
}
