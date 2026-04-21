package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// ======================= 测试lru-k ============================

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
	if lru.cachelistSub.Len() != 0 {
		t.Errorf("sub list should be empty after promotion")
	}
	if lru.cachelistMain.Len() != 1 {
		t.Errorf("main list should have 1 element")
	}
}

// ======================= 测试cache ============================

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := RegisterGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
