package core

import (
	"goProject/cache/util"
	"sync"
)

type cache struct {
	mutex      sync.Mutex
	lru        *LruCache
	cacheBytes int64
}

func (c *cache) add(key string, k int, value util.ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = NewCache(c.cacheBytes, k, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value util.ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(util.ByteView), ok
	}
	return
}
