package cache

import (
	"fmt"
	"log"
	"sync"
)

type cache struct {
	mutex      sync.Mutex
	lru        *LruCache
	cacheBytes int64
}

func (c *cache) add(key string, k int, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = NewCache(c.cacheBytes, k, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func RegisterGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("cache: getter is nil")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	defer mutex.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, 5, value)
}
