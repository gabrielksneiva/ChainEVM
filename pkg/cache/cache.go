package cache

import (
	"sync"
	"time"
)

// CacheEntry representa uma entrada de cache
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// RPCCache implementa um cache para chamadas RPC read-only
type RPCCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
	maxSize int
}

// NewRPCCache cria um novo cache RPC
func NewRPCCache(ttl time.Duration, maxSize int) *RPCCache {
	return &RPCCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get retorna um valor do cache se existir e não expirou
func (c *RPCCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	// Verificar se expirou
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set armazena um valor no cache
func (c *RPCCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Se cache está cheio, limpar entradas expiradas
	if len(c.entries) >= c.maxSize {
		c.evictExpired()
	}

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete remove uma entrada do cache
func (c *RPCCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear limpa todo o cache
func (c *RPCCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// evictExpired remove entradas expiradas (deve ser chamado com lock)
func (c *RPCCache) evictExpired() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Size retorna o número de entradas no cache
func (c *RPCCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
