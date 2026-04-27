package core

import "container/list"

// 底层lru-k
type LruCache struct {
	maxBytes      int64
	nbytes        int64
	k             int
	CachelistMain *list.List // >=k次的主缓存
	CachelistSub  *list.List // <k次
	cache         map[string]*list.Element
	history       map[string]int                // core-k次
	OnEvicted     func(key string, value Value) // 缓存被移除时的回调
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewCache(maxBytes int64, K int, onEvicted func(string, Value)) *LruCache {
	return &LruCache{
		maxBytes:      maxBytes,
		k:             K,
		CachelistMain: list.New(),
		CachelistSub:  list.New(),
		cache:         make(map[string]*list.Element),
		history:       make(map[string]int),
		OnEvicted:     onEvicted,
	}
}

func (c *LruCache) promote(key string, ele *list.Element) {
	// 1. 从 subList 移除
	c.CachelistSub.Remove(ele)

	// 2. 插入 mainList（创建新节点）
	newEle := c.CachelistMain.PushFront(ele.Value)

	// 3. 更新 cache 指针
	c.cache[key] = newEle
}

func (c *LruCache) Get(key string) (value Value, ok bool) {
	ele, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	c.history[key]++
	if c.history[key] == c.k {
		c.promote(key, ele) // 晋升
		ele = c.cache[key]
		c.CachelistMain.MoveToFront(ele)
	} else if c.history[key] > c.k {
		c.CachelistMain.MoveToFront(ele)
	} else {
		c.CachelistSub.MoveToFront(ele)
	}
	kv := ele.Value.(*entry)
	return kv.value, true
}

func (c *LruCache) delete(cacheList *list.List, ele *list.Element) {
	cacheList.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)
	delete(c.history, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *LruCache) Remove() {
	// 优先淘汰 subList（访问频率低）
	if ele := c.CachelistSub.Back(); ele != nil {
		c.delete(c.CachelistSub, ele)
		return
	}
	// subList 空了，再淘汰 mainList
	if ele := c.CachelistMain.Back(); ele != nil {
		c.delete(c.CachelistMain, ele)
	}
}

func (c *LruCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.history[key]++
		if c.history[key] == c.k {
			c.promote(key, ele)
			ele = c.cache[key]
			c.CachelistMain.MoveToFront(ele)
		} else if c.history[key] > c.k {
			c.CachelistMain.MoveToFront(ele)
		} else {
			c.CachelistSub.MoveToFront(ele)
		}
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.CachelistSub.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.history[key] = 1
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.Remove()
	}
}

func (c *LruCache) Len() int {
	return c.CachelistMain.Len() + c.CachelistSub.Len()
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
