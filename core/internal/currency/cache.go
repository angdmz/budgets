package currency

import (
	"sync"
	"time"

	"github.com/budgets/core/internal/domain"
)

type cacheKey struct {
	from domain.Currency
	to   domain.Currency
}

type cacheEntry struct {
	rate      ExchangeRate
	expiresAt time.Time
}

// InMemoryCache is a simple in-memory cache for exchange rates
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[cacheKey]cacheEntry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[cacheKey]cacheEntry),
	}
}

func (c *InMemoryCache) Get(from, to domain.Currency) (*ExchangeRate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey{from: from, to: to}
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return &entry.rate, true
}

func (c *InMemoryCache) Set(rate ExchangeRate, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{from: rate.FromCurrency, to: rate.ToCurrency}
	c.entries[key] = cacheEntry{
		rate:      rate,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[cacheKey]cacheEntry)
}

// Cleanup removes expired entries
func (c *InMemoryCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}
