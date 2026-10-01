package currency

import (
	"sync"
	"time"

	"github.com/budgets/core/internal/domain"
)

type cacheKey struct {
	from  domain.Currency
	to    domain.Currency
	quote domain.QuoteType
	// asOf is the rate date (YYYY-MM-DD) for historical rates, empty for
	// current rates.
	asOf string
}

func asOfKey(asOf *time.Time) string {
	if asOf == nil {
		return ""
	}
	return asOf.Format("2006-01-02")
}

type cacheEntry struct {
	rate      ExchangeRate
	expiresAt time.Time
}

// InMemoryCache is a simple in-memory cache for exchange rates.
// Current and historical rates share the store; the key distinguishes them
// by date.
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[cacheKey]cacheEntry
	now     func() time.Time
}

func NewInMemoryCache() *InMemoryCache {
	return NewInMemoryCacheWithClock(time.Now)
}

// NewInMemoryCacheWithClock builds a cache whose expiry is measured with the
// injected clock.
func NewInMemoryCacheWithClock(now func() time.Time) *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[cacheKey]cacheEntry),
		now:     now,
	}
}

func (c *InMemoryCache) Get(from, to domain.Currency, quote domain.QuoteType, asOf *time.Time) (*ExchangeRate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey{from: from, to: to, quote: quote, asOf: asOfKey(asOf)}
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if c.now().After(entry.expiresAt) {
		return nil, false
	}

	return &entry.rate, true
}

func (c *InMemoryCache) Set(rate ExchangeRate, asOf *time.Time, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{from: rate.FromCurrency, to: rate.ToCurrency, quote: rate.Quote, asOf: asOfKey(asOf)}
	c.entries[key] = cacheEntry{
		rate:      rate,
		expiresAt: c.now().Add(ttl),
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

	now := c.now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}
