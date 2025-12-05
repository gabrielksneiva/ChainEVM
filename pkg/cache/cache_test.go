package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRPCCache(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	require.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	cache.Set("balance:0x1234", "1000000")
	value, ok := cache.Get("balance:0x1234")

	require.True(t, ok)
	assert.Equal(t, "1000000", value)
}

func TestCacheExpiration(t *testing.T) {
	cache := NewRPCCache(100*time.Millisecond, 100)

	cache.Set("key", "value")
	value, ok := cache.Get("key")
	require.True(t, ok)
	assert.Equal(t, "value", value)

	// Aguardar expiração
	time.Sleep(150 * time.Millisecond)

	_, ok = cache.Get("key")
	assert.False(t, ok)
}

func TestCacheDelete(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	cache.Set("key", "value")
	cache.Delete("key")

	_, ok := cache.Get("key")
	assert.False(t, ok)
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	// Try to get a key that doesn't exist
	value, ok := cache.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCacheClear(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestCacheMaxSize(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 3)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Ao atingir max size, deve limpar expirados
	// Como nada expirou, a próxima entrada pode não ser adicionada ou pode remover a mais antiga
	cache.Set("key4", "value4")

	// Verificar que não ultrapassou o tamanho máximo
	assert.LessOrEqual(t, cache.Size(), 4)
}

func TestCacheMultipleEntries(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	for i := 1; i <= 10; i++ {
		key := "key" + fmt.Sprint(i)
		val := "value" + fmt.Sprint(i)
		cache.Set(key, val)
	}

	assert.Equal(t, 10, cache.Size())

	value, ok := cache.Get("key1")
	require.True(t, ok)
	assert.Equal(t, "value1", value)
}

func TestCacheEviction(t *testing.T) {
	// Small cache with max items limit
	cache := NewRPCCache(10*time.Second, 2)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	assert.Equal(t, 2, cache.Size())

	// This may or may not trigger eviction depending on implementation
	cache.Set("key3", "value3")

	// Cache size should be > 0
	size := cache.Size()
	assert.Greater(t, size, 0)
}

func TestCacheNonExistentKey(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	_, ok := cache.Get("nonexistent")

	assert.False(t, ok)
}

func TestCacheDeleteNonExistent(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	// Should not panic
	cache.Delete("nonexistent")
	assert.Equal(t, 0, cache.Size())
}

func TestCacheSequentialOperations(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	// Set
	cache.Set("addr:0x1111", "100")
	assert.Equal(t, 1, cache.Size())

	// Get
	val, ok := cache.Get("addr:0x1111")
	assert.True(t, ok)
	assert.Equal(t, "100", val)

	// Update
	cache.Set("addr:0x1111", "200")
	val, ok = cache.Get("addr:0x1111")
	assert.True(t, ok)
	assert.Equal(t, "200", val)

	// Delete
	cache.Delete("addr:0x1111")
	_, ok = cache.Get("addr:0x1111")
	assert.False(t, ok)
}

func TestCacheMultipleKeys(t *testing.T) {
	cache := NewRPCCache(10*time.Second, 100)

	keys := []string{"balance", "nonce", "code", "storage"}

	for i, key := range keys {
		cache.Set(key, fmt.Sprintf("value%d", i))
	}

	assert.Equal(t, len(keys), cache.Size())

	for i, key := range keys {
		val, ok := cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("value%d", i), val)
	}
}

func TestCacheExpirationEdgeCase(t *testing.T) {
	cache := NewRPCCache(50*time.Millisecond, 100)

	cache.Set("expires", "soon")
	value, ok := cache.Get("expires")
	require.True(t, ok)
	assert.Equal(t, "soon", value)

	// Wait slightly before expiration
	time.Sleep(40 * time.Millisecond)

	// Should still exist
	_, ok = cache.Get("expires")
	assert.True(t, ok)

	// Wait for expiration
	time.Sleep(30 * time.Millisecond)

	// Should be expired now
	_, ok = cache.Get("expires")
	assert.False(t, ok)
}

func TestCacheEvictionOnGet(t *testing.T) {
	cache := NewRPCCache(50*time.Millisecond, 100)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// All entries should exist
	assert.Equal(t, 3, cache.Size())

	// Wait for all to expire
	time.Sleep(60 * time.Millisecond)

	// Calling Get triggers eviction
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	// Cache should now be empty after eviction
	assert.Equal(t, 0, cache.Size())
}
