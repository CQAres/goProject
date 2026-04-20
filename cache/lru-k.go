package lru

import "container/list"

type Cache struct {
	maxBytes      int64
	nbytes        int64
	k             int
	cachelistMain *list.List // >=k次的主缓存
	cachelistSub  *list.List // <k次
	cache         map[string]*list.Element
	history       map[string]int                // lru-k次
	OnEvicted     func(key string, value Value) // 缓存被移除时的回调
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewCache(maxBytes int64, K int, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:      maxBytes,
		k:             K,
		cachelistMain: list.New(),
		cachelistSub:  list.New(),
		cache:         make(map[string]*list.Element),
		history:       make(map[string]int),
		OnEvicted:     onEvicted,
	}
}

func (c *Cache) promote(key string, ele *list.Element) {
	// 1. 从 subList 移除
	c.cachelistSub.Remove(ele)

	// 2. 插入 mainList（创建新节点）
	newEle := c.cachelistMain.PushFront(ele.Value)

	// 3. 更新 cache 指针
	c.cache[key] = newEle
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	ele, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	c.history[key]++
	if c.history[key] == c.k {
		c.promote(key, ele) // 晋升
		ele = c.cache[key]
		c.cachelistMain.MoveToFront(ele)
	} else if c.history[key] > c.k {
		c.cachelistMain.MoveToFront(ele)
	} else {
		c.cachelistSub.MoveToFront(ele)
	}
	kv := ele.Value.(*entry)
	return kv.value, true
}

func (c *Cache) delete(cacheList *list.List, ele *list.Element) {
	cacheList.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)
	delete(c.history, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Remove() {
	// 优先淘汰 subList（访问频率低）
	if ele := c.cachelistSub.Back(); ele != nil {
		c.delete(c.cachelistSub, ele)
		return
	}
	// subList 空了，再淘汰 mainList
	if ele := c.cachelistMain.Back(); ele != nil {
		c.delete(c.cachelistMain, ele)
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.history[key]++
		if c.history[key] == c.k {
			c.promote(key, ele)
			ele = c.cache[key]
			c.cachelistMain.MoveToFront(ele)
		} else if c.history[key] > c.k {
			c.cachelistMain.MoveToFront(ele)
		} else {
			c.cachelistSub.MoveToFront(ele)
		}
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.cachelistSub.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.history[key] = 1
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.Remove()
	}
}

func (c *Cache) Len() int {
	return c.cachelistMain.Len() + c.cachelistSub.Len()
}
