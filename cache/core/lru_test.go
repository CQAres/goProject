package core

import (
	"reflect"
	"testing"
)

func TestLRUGet(t *testing.T) {
	// K=1 表示普通 LRU（所有 key 都在 mainList）
	lru := NewCache(0, 1, nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"

	// 计算前两个 key-value 的总字节数（作为容量上限）
	capacity := int64(len(k1) + len(v1) + len(k2) + len(v2))

	lru := NewCache(capacity, 1, nil) // K=1，普通 LRU

	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	// 此时缓存已满

	// 添加第三个，会淘汰最久未使用的（k1）
	lru.Add(k3, String(v3))

	// k1 应该被淘汰
	if _, ok := lru.Get(k1); ok {
		t.Fatalf("expected key1 to be evicted, but it's still in cache")
	}

	// 缓存中应只有 k2 和 k3
	if lru.Len() != 2 {
		t.Fatalf("expected cache length 2, got %d", lru.Len())
	}

	// 确保 k2 和 k3 还在
	if _, ok := lru.Get(k2); !ok {
		t.Fatalf("key2 should still be in cache")
	}
	if _, ok := lru.Get(k3); !ok {
		t.Fatalf("key3 should still be in cache")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	// 设置容量为 10 字节
	lru := NewCache(10, 1, callback) // K=1

	// "key1" + "123456" = 4 + 6 = 10 → 刚好填满
	lru.Add("key1", String("123456")) // 4+6=10

	// 添加 "k2"（2+2=4），总需求 14 > 10 → 淘汰 key1
	lru.Add("k2", String("k2")) // 2+2=4

	// 添加 "k3"（2+2=4），总需求 8 <= 10？不，当前是 k2(4) + k3(4)=8，OK
	lru.Add("k3", String("k3"))

	// 添加 "k4"（2+2=4），总需求 12 > 10 → 淘汰 k2（最久未用）
	lru.Add("k4", String("k4"))

	// 所以被驱逐的是：key1, k2
	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("OnEvicted called with %v, expected %v", keys, expect)
	}
}

func TestLRUKPromote(t *testing.T) {
	evicted := make([]string, 0)
	lru := NewCache(100, 2, func(k string, v Value) {
		evicted = append(evicted, k)
	})

	lru.Add("a", String("x"))
	lru.Get("a") // 第1次访问（history=1，仍在 sub）
	lru.Get("a") // 第2次访问（history=2 == K，晋升到 main）

	// 此时 "a" 应在 mainList
	if lru.CachelistSub.Len() != 0 {
		t.Errorf("sub list should be empty after promotion")
	}
	if lru.CachelistMain.Len() != 1 {
		t.Errorf("main list should have 1 element")
	}
}

// 测试LRU-K场景：多次访问后晋升
func TestLRUKMultipleAccess(t *testing.T) {
	lru := NewCache(100, 2, nil) // K=2，需要访问2次才晋升

	lru.Add("key", String("value"))

	// 初始在subList
	if lru.CachelistSub.Len() != 1 {
		t.Errorf("sub list should have 1 element after Add, got %d", lru.CachelistSub.Len())
	}
	if lru.CachelistMain.Len() != 0 {
		t.Errorf("main list should be empty, got %d", lru.CachelistMain.Len())
	}

	lru.Get("key") // 第1次访问，history=2，晋升到main

	// 应在mainList
	if lru.CachelistMain.Len() != 1 {
		t.Errorf("main list should have 1 element after promotion, got %d", lru.CachelistMain.Len())
	}
	if lru.CachelistSub.Len() != 0 {
		t.Errorf("sub list should be empty after promotion, got %d", lru.CachelistSub.Len())
	}
}

// 测试LRU-K Add时晋升
func TestLRUKAddPromotion(t *testing.T) {
	lru := NewCache(100, 2, nil) // K=2

	// Add两次同一key，应该晋升
	lru.Add("key", String("v1"))
	lru.Add("key", String("v2")) // history达到2，晋升

	if lru.CachelistMain.Len() != 1 {
		t.Errorf("should be in main list after 2 Adds, got main=%d", lru.CachelistMain.Len())
	}
}

// 测试更新已存在的key
func TestLRUUpdate(t *testing.T) {
	lru := NewCache(100, 1, nil)

	lru.Add("key", String("value1"))
	if v, ok := lru.Get("key"); !ok || string(v.(String)) != "value1" {
		t.Errorf("first value should be value1")
	}

	lru.Add("key", String("value2"))
	if v, ok := lru.Get("key"); !ok || string(v.(String)) != "value2" {
		t.Errorf("updated value should be value2, got %v", v)
	}
}

// 测试容量为0时不淘汰
func TestLRUZeroCapacity(t *testing.T) {
	lru := NewCache(0, 1, nil)

	lru.Add("key1", String("value1"))
	lru.Add("key2", String("value2"))

	// 容量为0时不应淘汰任何元素
	if lru.Len() != 2 {
		t.Errorf("expected length 2, got %d", lru.Len())
	}
}

// 测试Remove操作
func TestLRURemove(t *testing.T) {
	lru := NewCache(100, 1, nil)

	lru.Add("key1", String("value1"))
	lru.Add("key2", String("value2"))

	removed := false
	lru.OnEvicted = func(key string, value Value) {
		removed = true
	}

	lru.Remove()

	if !removed {
		t.Errorf("OnEvicted should be called on Remove")
	}
	if lru.Len() != 1 {
		t.Errorf("expected length 1 after remove, got %d", lru.Len())
	}
}

// 测试大量数据淘汰
func TestLRUEvictMany(t *testing.T) {
	evicted := make([]string, 0)
	lru := NewCache(50, 1, func(k string, v Value) {
		evicted = append(evicted, k)
	})

	// 添加10个key，每个约10字节，总共约100字节
	for i := 0; i < 10; i++ {
		lru.Add(string(rune('a'+i)), String("value123"))
	}

	// 应该淘汰多个key
	if len(evicted) == 0 {
		t.Errorf("some keys should be evicted")
	}

	if lru.Len() > 5 {
		t.Errorf("cache should not exceed capacity, got %d", lru.Len())
	}
}

// 测试空缓存Get
func TestLRUEmptyGet(t *testing.T) {
	lru := NewCache(100, 1, nil)

	if _, ok := lru.Get("notexist"); ok {
		t.Errorf("should return false for non-existent key")
	}
}
