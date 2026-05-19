package lru

import (
	"time"
)

func (c *Cache[K, V]) janitor(interval time.Duration) {
	ticker, stop := c.clock.Ticker(interval)
	defer stop()

	for {
		select {
		case <-c.stopJanitor:
			return
		case <-ticker:
			c.deleteExpired()
		}
	}
}

func (c *Cache[K, V]) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	for key, elem := range c.items {
		if !elem.loading && elem.pinCount == 0 && c.isExpired(elem) {
			c.evictElement(key, elem)
		}
	}
}
