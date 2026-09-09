package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

type CacheItem struct {
	Value any
	Exp   time.Time
}

type InMemoryCache struct {
	mu   sync.RWMutex
	data map[string]CacheItem
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{data: make(map[string]CacheItem)}
}

func (c *InMemoryCache) Set(k string, v any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := time.Now().UTC().Add(ttl)
	c.data[k] = CacheItem{Value: v, Exp: exp}
}

func (c *InMemoryCache) Get(k string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[k]
	if !ok {
		return nil, false
	}

	if !item.Exp.IsZero() && time.Now().UTC().After(item.Exp) {
		return nil, false
	}

	return item.Value, true
}

func (c *InMemoryCache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, k)
}

func (c *InMemoryCache) DeleteByPrefix(prefix string) {
	for k := range c.data {
		if strings.HasPrefix(k, prefix) {
			c.mu.Lock()
			defer c.mu.Unlock()
			delete(c.data, k)
		}
	}
}

func (c *InMemoryCache) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				now := time.Now().UTC()

				c.mu.Lock()
				for k, item := range c.data {
					if item.Exp.IsZero() && now.After(item.Exp) {
						delete(c.data, k)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
}
