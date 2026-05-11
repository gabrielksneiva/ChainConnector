package monitor

import (
	"sync"
)

type BloomFilterCache struct {
	mu    sync.RWMutex
	items map[string]struct{}
}

func NewBloomFilterCache() *BloomFilterCache {
	return &BloomFilterCache{
		items: make(map[string]struct{}),
	}
}

func (b *BloomFilterCache) AddString(value string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[value] = struct{}{}
}

func (b *BloomFilterCache) TestString(value string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.items[value]
	return ok
}
