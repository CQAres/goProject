package core

import (
	"goProject/cache/util"
	"sync"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// 测试cache的add和get
func TestCacheAddAndGet(t *testing.T) {
	c := &cache{
		lru:        NewCache(100, 1, nil),
		cacheBytes: 100,
	}

	c.add("key1", 1, util.ByteView{B: []byte("value1")})
	c.add("key2", 1, util.ByteView{B: []byte("value2")})

	v, ok := c.get("key1")
	if !ok || v.String() != "value1" {
		t.Errorf("failed to get key1, got %v, ok=%v", v, ok)
	}

	v, ok = c.get("key2")
	if !ok || v.String() != "value2" {
		t.Errorf("failed to get key2, got %v, ok=%v", v, ok)
	}
}

// 测试空cache的get
func TestCacheEmptyGet(t *testing.T) {
	c := &cache{
		lru:        nil,
		cacheBytes: 100,
	}

	v, ok := c.get("notexist")
	if ok || v.B != nil {
		t.Errorf("empty cache should return false")
	}
}

// 测试cache容量满时的淘汰
func TestCacheEviction(t *testing.T) {
	c := &cache{
		lru:        NewCache(10, 1, nil),
		cacheBytes: 10,
	}

	evicted := make([]string, 0)
	c.lru.OnEvicted = func(key string, value Value) {
		evicted = append(evicted, key)
	}

	c.add("key1", 1, util.ByteView{B: []byte("1234")}) // 4+4=8
	c.add("key2", 1, util.ByteView{B: []byte("5678")}) // 4+4=8, 总16 > 10，会淘汰key1

	if len(evicted) == 0 {
		t.Errorf("key1 should be evicted")
	}
	if evicted[0] != "key1" {
		t.Errorf("expected key1 to be evicted first, got %s", evicted[0])
	}
}

// 测试并发访问cache
func TestCacheConcurrentAccess(t *testing.T) {
	c := &cache{
		lru:        NewCache(10000, 1, nil),
		cacheBytes: 10000,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune('a'+id%26))
			c.add(key, 1, util.ByteView{B: []byte("value")})
			c.get(key)
		}(i)
	}
	wg.Wait()

	if c.lru.Len() == 0 {
		t.Errorf("cache should have some elements")
	}
}

// 测试cache更新已存在的key
func TestCacheUpdateKey(t *testing.T) {
	c := &cache{
		lru:        NewCache(100, 1, nil),
		cacheBytes: 100,
	}

	c.add("key", 1, util.ByteView{B: []byte("value1")})
	c.add("key", 1, util.ByteView{B: []byte("value2")}) // 更新

	v, ok := c.get("key")
	if !ok || v.String() != "value2" {
		t.Errorf("key should have updated value, got %v", v)
	}

	// 长度应该还是1
	if c.lru.Len() != 1 {
		t.Errorf("cache length should be 1, got %d", c.lru.Len())
	}
}
